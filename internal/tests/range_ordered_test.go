package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// testNumeric is a float64-backed type that implements typx.Ordered[testNumeric].
// MarshalJSON/UnmarshalJSON handle JSON encoding (used when stored in jsonb/text columns).
type testNumeric struct{ v float64 }

func (n testNumeric) Compare(other testNumeric) int {
	if n.v < other.v {
		return -1
	}
	if n.v > other.v {
		return 1
	}
	return 0
}

// MarshalJSON encodes testNumeric as a JSON number.
func (n testNumeric) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(n.v, 'f', -1, 64)), nil
}

// UnmarshalJSON parses a JSON number into testNumeric.
func (n *testNumeric) UnmarshalJSON(data []byte) error {
	v, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	n.v = v
	return nil
}

// MarshalBSONValue encodes testNumeric as a BSON double.
func (n testNumeric) MarshalBSONValue() (byte, []byte, error) {
	t, data, err := bson.MarshalValue(n.v)
	return byte(t), data, err
}

// UnmarshalBSONValue decodes a BSON double into testNumeric.
func (n *testNumeric) UnmarshalBSONValue(t byte, data []byte) error {
	return bson.UnmarshalValue(bson.Type(t), data, &n.v)
}

func num(v float64) testNumeric { return testNumeric{v} }

func TestOrderedRange_Scan_Value(t *testing.T) {
	ctx := context.Background()
	db := postgresClient

	_, err := db.ExecContext(ctx, `CREATE TABLE ordered_range_val (id serial PRIMARY KEY, val jsonb NOT NULL)`)
	require.NoError(t, err)

	tests := []struct {
		name string
		rng  typx.OrderedRange[testNumeric]
	}{
		{
			name: "closed-closed positive",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(1.5)},
				Upper: typx.OrderedBound[testNumeric]{Val: num(10.5)},
			},
		},
		{
			name: "open-closed",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(-5.0), Exclusive: true},
				Upper: typx.OrderedBound[testNumeric]{Val: num(5.0)},
			},
		},
		{
			name: "closed-open",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(-100.25)},
				Upper: typx.OrderedBound[testNumeric]{Val: num(-0.5), Exclusive: true},
			},
		},
		{
			name: "open-open",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(0), Exclusive: true},
				Upper: typx.OrderedBound[testNumeric]{Val: num(1000), Exclusive: true},
			},
		},
		{
			name: "unbounded lower",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Unbounded: true},
				Upper: typx.OrderedBound[testNumeric]{Val: num(5.0)},
			},
		},
		{
			name: "unbounded upper",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(1.0)},
				Upper: typx.OrderedBound[testNumeric]{Unbounded: true},
			},
		},
		{
			name: "both unbounded",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Unbounded: true},
				Upper: typx.OrderedBound[testNumeric]{Unbounded: true},
			},
		},
		{
			name: "exclusive lower unbounded upper",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(3.0), Exclusive: true},
				Upper: typx.OrderedBound[testNumeric]{Unbounded: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.rng.Value()
			require.NoError(t, err)
			require.NotNil(t, val)

			_, err = db.ExecContext(ctx, `INSERT INTO ordered_range_val (val) VALUES ($1)`, val)
			require.NoError(t, err)

			var scanned typx.OrderedRange[testNumeric]
			row := db.QueryRowContext(ctx, `SELECT val FROM ordered_range_val ORDER BY id DESC LIMIT 1`)
			err = row.Scan(&scanned)
			require.NoError(t, err)

			// jsonb round-trip: bound types and values are preserved exactly.
			require.Equal(t, tt.rng, scanned)
		})
	}
}

func TestOrderedMultiRange_Scan_Value(t *testing.T) {
	ctx := context.Background()
	db := postgresClient

	_, err := db.ExecContext(ctx, `CREATE TABLE ordered_multi_range_val (id serial PRIMARY KEY, val jsonb NOT NULL)`)
	require.NoError(t, err)

	tests := []struct {
		name string
		mr   typx.OrderedMultiRange[testNumeric]
	}{
		{
			name: "empty",
			mr:   typx.OrderedMultiRange[testNumeric]{},
		},
		{
			name: "single closed-closed",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(1.5)}, Upper: typx.OrderedBound[testNumeric]{Val: num(10.5)}},
			},
		},
		{
			name: "two non-overlapping with mixed bounds",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(1.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(5.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(7.0), Exclusive: true}, Upper: typx.OrderedBound[testNumeric]{Val: num(10.0)}},
			},
		},
		{
			name: "three ranges",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(-20.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(-10.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(0.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(5.5)}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(100.0), Exclusive: true}, Upper: typx.OrderedBound[testNumeric]{Val: num(200.0)}},
			},
		},
		{
			name: "unbounded lower bound",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Unbounded: true}, Upper: typx.OrderedBound[testNumeric]{Val: num(5.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(7.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(10.0)}},
			},
		},
		{
			name: "unbounded upper bound",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(-5.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(0.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(3.0)}, Upper: typx.OrderedBound[testNumeric]{Unbounded: true}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.mr.Value()
			require.NoError(t, err)

			_, err = db.ExecContext(ctx, `INSERT INTO ordered_multi_range_val (val) VALUES ($1)`, val)
			require.NoError(t, err)

			var scanned typx.OrderedMultiRange[testNumeric]
			row := db.QueryRowContext(ctx, `SELECT val FROM ordered_multi_range_val ORDER BY id DESC LIMIT 1`)
			err = row.Scan(&scanned)
			require.NoError(t, err)

			// jsonb round-trip: bound types and values are preserved exactly.
			require.Equal(t, tt.mr, scanned)
		})
	}
}

func TestOrderedRange_MarshalUnmarshal_Mongo(t *testing.T) {
	ctx := context.Background()
	db := mongoClient.Database("ordered_range_mongo_test")

	coll := db.Collection("ordered_ranges")

	tests := []struct {
		name string
		rng  typx.OrderedRange[testNumeric]
	}{
		{
			name: "closed-closed positive",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(1.5)},
				Upper: typx.OrderedBound[testNumeric]{Val: num(10.5)},
			},
		},
		{
			name: "open-closed",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(-5.0), Exclusive: true},
				Upper: typx.OrderedBound[testNumeric]{Val: num(5.0)},
			},
		},
		{
			name: "closed-open",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(-100.25)},
				Upper: typx.OrderedBound[testNumeric]{Val: num(-0.5), Exclusive: true},
			},
		},
		{
			name: "open-open",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(0), Exclusive: true},
				Upper: typx.OrderedBound[testNumeric]{Val: num(1000), Exclusive: true},
			},
		},
		{
			name: "unbounded lower",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Unbounded: true},
				Upper: typx.OrderedBound[testNumeric]{Val: num(5.0)},
			},
		},
		{
			name: "unbounded upper",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(1.0)},
				Upper: typx.OrderedBound[testNumeric]{Unbounded: true},
			},
		},
		{
			name: "both unbounded",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Unbounded: true},
				Upper: typx.OrderedBound[testNumeric]{Unbounded: true},
			},
		},
		{
			name: "exclusive lower unbounded upper",
			rng: typx.OrderedRange[testNumeric]{
				Lower: typx.OrderedBound[testNumeric]{Val: num(3.0), Exclusive: true},
				Upper: typx.OrderedBound[testNumeric]{Unbounded: true},
			},
		},
	}

	type doc struct {
		Val typx.OrderedRange[testNumeric] `bson:"val"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := coll.InsertOne(ctx, doc{Val: tt.rng})
			require.NoError(t, err)

			var out doc
			err = coll.FindOne(ctx, bson.M{"_id": res.InsertedID}).Decode(&out)
			require.NoError(t, err)

			require.Equal(t, tt.rng, out.Val)
		})
	}
}

func TestOrderedMultiRange_MarshalUnmarshal_Mongo(t *testing.T) {
	ctx := context.Background()
	db := mongoClient.Database("ordered_multi_range_mongo_test")

	coll := db.Collection("ordered_multiranges")

	tests := []struct {
		name string
		mr   typx.OrderedMultiRange[testNumeric]
	}{
		{
			name: "empty",
			mr:   typx.OrderedMultiRange[testNumeric]{},
		},
		{
			name: "single closed-closed",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(1.5)}, Upper: typx.OrderedBound[testNumeric]{Val: num(10.5)}},
			},
		},
		{
			name: "two non-overlapping with mixed bounds",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(1.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(5.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(7.0), Exclusive: true}, Upper: typx.OrderedBound[testNumeric]{Val: num(10.0)}},
			},
		},
		{
			name: "three ranges",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(-20.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(-10.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(0.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(5.5)}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(100.0), Exclusive: true}, Upper: typx.OrderedBound[testNumeric]{Val: num(200.0)}},
			},
		},
		{
			name: "unbounded lower bound",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Unbounded: true}, Upper: typx.OrderedBound[testNumeric]{Val: num(5.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(7.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(10.0)}},
			},
		},
		{
			name: "unbounded upper bound",
			mr: typx.OrderedMultiRange[testNumeric]{
				{Lower: typx.OrderedBound[testNumeric]{Val: num(-5.0)}, Upper: typx.OrderedBound[testNumeric]{Val: num(0.0), Exclusive: true}},
				{Lower: typx.OrderedBound[testNumeric]{Val: num(3.0)}, Upper: typx.OrderedBound[testNumeric]{Unbounded: true}},
			},
		},
	}

	type doc struct {
		Val typx.OrderedMultiRange[testNumeric] `bson:"val"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := coll.InsertOne(ctx, doc{Val: tt.mr})
			require.NoError(t, err)

			var out doc
			err = coll.FindOne(ctx, bson.M{"_id": res.InsertedID}).Decode(&out)
			require.NoError(t, err)

			require.Equal(t, tt.mr, out.Val)
		})
	}
}
func TestOrderedRange_Scan_Value_Time(t *testing.T) {
	ctx := context.Background()
	db := postgresClient

	_, err := db.ExecContext(ctx, `CREATE TABLE ordered_range_time_val (id serial PRIMARY KEY, val tstzrange NOT NULL)`)
	require.NoError(t, err)

	tests := []struct {
		name string
		rng  typx.OrderedRange[typx.DateTime]
	}{
		{
			name: "closed-closed",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-01-01T00:00:00Z"))}},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-12-31T23:59:59Z"))}},
			},
		},
		{
			name: "open-closed",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2020-06-15T12:00:00Z"))}, Exclusive: true},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2025-06-15T12:00:00Z"))}},
			},
		},
		{
			name: "closed-open",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2000-01-01T00:00:00Z"))}},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2001-01-01T00:00:00Z"))}, Exclusive: true},
			},
		},
		{
			name: "open-open",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2023-03-01T08:30:00Z"))}, Exclusive: true},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2023-09-01T08:30:00Z"))}, Exclusive: true},
			},
		},
		{
			name: "unbounded lower",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Unbounded: true},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-06-01T00:00:00Z"))}},
			},
		},
		{
			name: "unbounded upper",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-06-01T00:00:00Z"))}},
				Upper: typx.OrderedBound[typx.DateTime]{Unbounded: true},
			},
		},
		{
			name: "both unbounded",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Unbounded: true},
				Upper: typx.OrderedBound[typx.DateTime]{Unbounded: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.rng.Value()
			require.NoError(t, err)
			require.NotNil(t, val)

			_, err = db.ExecContext(ctx, `INSERT INTO ordered_range_time_val (val) VALUES ($1)`, val)
			require.NoError(t, err)

			var scanned typx.OrderedRange[typx.DateTime]
			row := db.QueryRowContext(ctx, `SELECT val FROM ordered_range_time_val ORDER BY id DESC LIMIT 1`)
			err = row.Scan(&scanned)
			require.NoError(t, err)

			require.Equal(t, tt.rng, scanned)
		})
	}
}

func TestOrderedMultiRange_Scan_Value_Time(t *testing.T) {
	ctx := context.Background()
	db := postgresClient

	_, err := db.ExecContext(ctx, `CREATE TABLE ordered_multi_range_time_val (id serial PRIMARY KEY, val tstzmultirange NOT NULL)`)
	require.NoError(t, err)

	tests := []struct {
		name string
		mr   typx.OrderedMultiRange[typx.DateTime]
	}{
		{
			name: "empty",
			mr:   typx.OrderedMultiRange[typx.DateTime]{},
		},
		{
			name: "single closed-closed",
			mr: typx.OrderedMultiRange[typx.DateTime]{
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-06-30T23:59:59Z"))}}},
			},
		},
		{
			name: "two non-overlapping",
			mr: typx.OrderedMultiRange[typx.DateTime]{
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2022-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2022-06-30T00:00:00Z"))}, Exclusive: true}},
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2023-01-01T00:00:00Z"))}, Exclusive: true}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2023-12-31T00:00:00Z"))}}},
			},
		},
		{
			name: "unbounded lower",
			mr: typx.OrderedMultiRange[typx.DateTime]{
				{Lower: typx.OrderedBound[typx.DateTime]{Unbounded: true}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2020-01-01T00:00:00Z"))}, Exclusive: true}},
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2021-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2021-12-31T00:00:00Z"))}}},
			},
		},
		{
			name: "unbounded upper",
			mr: typx.OrderedMultiRange[typx.DateTime]{
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2019-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2019-06-01T00:00:00Z"))}, Exclusive: true}},
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2025-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Unbounded: true}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.mr.Value()
			require.NoError(t, err)

			_, err = db.ExecContext(ctx, `INSERT INTO ordered_multi_range_time_val (val) VALUES ($1)`, val)
			require.NoError(t, err)

			var scanned typx.OrderedMultiRange[typx.DateTime]
			row := db.QueryRowContext(ctx, `SELECT val FROM ordered_multi_range_time_val ORDER BY id DESC LIMIT 1`)
			err = row.Scan(&scanned)
			require.NoError(t, err)

			require.Equal(t, tt.mr, scanned)
		})
	}
}

func TestOrderedRange_MarshalUnmarshal_Mongo_Time(t *testing.T) {
	ctx := context.Background()
	db := mongoClient.Database("ordered_range_time_mongo_test")

	coll := db.Collection("ordered_ranges_time")

	tests := []struct {
		name string
		rng  typx.OrderedRange[typx.DateTime]
	}{
		{
			name: "closed-closed",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-01-01T00:00:00Z"))}},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-12-31T23:59:59Z"))}},
			},
		},
		{
			name: "open-closed",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2020-06-15T12:00:00Z"))}, Exclusive: true},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2025-06-15T12:00:00Z"))}},
			},
		},
		{
			name: "unbounded lower",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Unbounded: true},
				Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-06-01T00:00:00Z"))}},
			},
		},
		{
			name: "unbounded upper",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-06-01T00:00:00Z"))}},
				Upper: typx.OrderedBound[typx.DateTime]{Unbounded: true},
			},
		},
		{
			name: "both unbounded",
			rng: typx.OrderedRange[typx.DateTime]{
				Lower: typx.OrderedBound[typx.DateTime]{Unbounded: true},
				Upper: typx.OrderedBound[typx.DateTime]{Unbounded: true},
			},
		},
	}

	type doc struct {
		Val typx.OrderedRange[typx.DateTime] `bson:"val"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := coll.InsertOne(ctx, doc{Val: tt.rng})
			require.NoError(t, err)

			var out doc
			err = coll.FindOne(ctx, bson.M{"_id": res.InsertedID}).Decode(&out)
			require.NoError(t, err)

			require.Equal(t, tt.rng, out.Val)
		})
	}
}

func TestOrderedMultiRange_MarshalUnmarshal_Mongo_Time(t *testing.T) {
	ctx := context.Background()
	db := mongoClient.Database("ordered_multi_range_time_mongo_test")

	coll := db.Collection("ordered_multiranges_time")

	tests := []struct {
		name string
		mr   typx.OrderedMultiRange[typx.DateTime]
	}{
		{
			name: "empty",
			mr:   typx.OrderedMultiRange[typx.DateTime]{},
		},
		{
			name: "single closed-closed",
			mr: typx.OrderedMultiRange[typx.DateTime]{
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2024-06-30T23:59:59Z"))}}},
			},
		},
		{
			name: "two non-overlapping",
			mr: typx.OrderedMultiRange[typx.DateTime]{
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2022-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2022-06-30T00:00:00Z"))}, Exclusive: true}},
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2023-01-01T00:00:00Z"))}, Exclusive: true}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2023-12-31T00:00:00Z"))}}},
			},
		},
		{
			name: "unbounded upper",
			mr: typx.OrderedMultiRange[typx.DateTime]{
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2019-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2019-06-01T00:00:00Z"))}, Exclusive: true}},
				{Lower: typx.OrderedBound[typx.DateTime]{Val: typx.DateTime{Time: typx.Must(time.Parse(time.RFC3339, "2025-01-01T00:00:00Z"))}}, Upper: typx.OrderedBound[typx.DateTime]{Unbounded: true}},
			},
		},
	}

	type doc struct {
		Val typx.OrderedMultiRange[typx.DateTime] `bson:"val"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := coll.InsertOne(ctx, doc{Val: tt.mr})
			require.NoError(t, err)

			var out doc
			err = coll.FindOne(ctx, bson.M{"_id": res.InsertedID}).Decode(&out)
			require.NoError(t, err)

			require.Equal(t, tt.mr, out.Val)
		})
	}
}
