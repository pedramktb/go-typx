package tests

import (
	"context"
	"testing"

	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRange_Scan_Value(t *testing.T) {
	ctx := context.Background()
	db := postgresClient

	_, err := db.ExecContext(ctx, `CREATE TABLE range_val (id serial PRIMARY KEY, val numrange NOT NULL)`)
	require.NoError(t, err)

	tests := []struct {
		name string
		rng  typx.Range[float64]
	}{
		{
			name: "closed-closed positive",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 1.5},
				Upper: typx.Bound[float64]{Val: 10.5},
			},
		},
		{
			name: "open-closed",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: -5.0, Exclusive: true},
				Upper: typx.Bound[float64]{Val: 5.0},
			},
		},
		{
			name: "closed-open",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: -100.25},
				Upper: typx.Bound[float64]{Val: -0.5, Exclusive: true},
			},
		},
		{
			name: "open-open",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 0, Exclusive: true},
				Upper: typx.Bound[float64]{Val: 1e9, Exclusive: true},
			},
		},
		{
			name: "unbounded lower",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Unbounded: true},
				Upper: typx.Bound[float64]{Val: 5.0},
			},
		},
		{
			name: "unbounded upper",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 1.0},
				Upper: typx.Bound[float64]{Unbounded: true},
			},
		},
		{
			name: "both unbounded",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Unbounded: true},
				Upper: typx.Bound[float64]{Unbounded: true},
			},
		},
		{
			name: "exclusive lower unbounded upper",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 3.0, Exclusive: true},
				Upper: typx.Bound[float64]{Unbounded: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.rng.Value()
			require.NoError(t, err)
			require.NotNil(t, val)

			_, err = db.ExecContext(ctx, `INSERT INTO range_val (val) VALUES ($1)`, val)
			require.NoError(t, err)

			var scanned typx.Range[float64]
			row := db.QueryRowContext(ctx, `SELECT val FROM range_val ORDER BY id DESC LIMIT 1`)
			err = row.Scan(&scanned)
			require.NoError(t, err)

			// numrange is continuous: PostgreSQL preserves bound types exactly.
			require.Equal(t, tt.rng, scanned)
		})
	}
}

func TestMultiRange_Scan_Value(t *testing.T) {
	ctx := context.Background()
	db := postgresClient

	_, err := db.ExecContext(ctx, `CREATE TABLE multirange_val (id serial PRIMARY KEY, val nummultirange NOT NULL)`)
	require.NoError(t, err)

	tests := []struct {
		name string
		mr   typx.MultiRange[float64]
	}{
		{
			name: "empty",
			mr:   typx.MultiRange[float64]{},
		},
		{
			name: "single closed-closed",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: 1.5}, Upper: typx.Bound[float64]{Val: 10.5}},
			},
		},
		{
			name: "two non-overlapping with mixed bounds",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: 1.0}, Upper: typx.Bound[float64]{Val: 5.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 7.0, Exclusive: true}, Upper: typx.Bound[float64]{Val: 10.0}},
			},
		},
		{
			name: "three ranges",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: -20.0}, Upper: typx.Bound[float64]{Val: -10.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 0.0}, Upper: typx.Bound[float64]{Val: 5.5}},
				{Lower: typx.Bound[float64]{Val: 100.0, Exclusive: true}, Upper: typx.Bound[float64]{Val: 200.0}},
			},
		},
		{
			name: "unbounded lower bound",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Unbounded: true}, Upper: typx.Bound[float64]{Val: 5.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 7.0}, Upper: typx.Bound[float64]{Val: 10.0}},
			},
		},
		{
			name: "unbounded upper bound",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: -5.0}, Upper: typx.Bound[float64]{Val: 0.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 3.0}, Upper: typx.Bound[float64]{Unbounded: true}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.mr.Value()
			require.NoError(t, err)

			_, err = db.ExecContext(ctx, `INSERT INTO multirange_val (val) VALUES ($1)`, val)
			require.NoError(t, err)

			var scanned typx.MultiRange[float64]
			row := db.QueryRowContext(ctx, `SELECT val FROM multirange_val ORDER BY id DESC LIMIT 1`)
			err = row.Scan(&scanned)
			require.NoError(t, err)

			// nummultirange is continuous: bound types and values are preserved exactly.
			require.Equal(t, tt.mr, scanned)
		})
	}
}

func TestRange_MarshalUnmarshal_Mongo(t *testing.T) {
	ctx := context.Background()
	db := mongoClient.Database("range_mongo_test")

	coll := db.Collection("ranges")

	tests := []struct {
		name string
		rng  typx.Range[float64]
	}{
		{
			name: "closed-closed positive",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 1.5},
				Upper: typx.Bound[float64]{Val: 10.5},
			},
		},
		{
			name: "open-closed",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: -5.0, Exclusive: true},
				Upper: typx.Bound[float64]{Val: 5.0},
			},
		},
		{
			name: "closed-open",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: -100.25},
				Upper: typx.Bound[float64]{Val: -0.5, Exclusive: true},
			},
		},
		{
			name: "open-open",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 0, Exclusive: true},
				Upper: typx.Bound[float64]{Val: 1e9, Exclusive: true},
			},
		},
		{
			name: "unbounded lower",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Unbounded: true},
				Upper: typx.Bound[float64]{Val: 5.0},
			},
		},
		{
			name: "unbounded upper",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 1.0},
				Upper: typx.Bound[float64]{Unbounded: true},
			},
		},
		{
			name: "both unbounded",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Unbounded: true},
				Upper: typx.Bound[float64]{Unbounded: true},
			},
		},
		{
			name: "exclusive lower unbounded upper",
			rng: typx.Range[float64]{
				Lower: typx.Bound[float64]{Val: 3.0, Exclusive: true},
				Upper: typx.Bound[float64]{Unbounded: true},
			},
		},
	}

	type doc struct {
		Val typx.Range[float64] `bson:"val"`
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

func TestMultiRange_MarshalUnmarshal_Mongo(t *testing.T) {
	ctx := context.Background()
	db := mongoClient.Database("multirange_mongo_test")

	coll := db.Collection("multiranges")

	tests := []struct {
		name string
		mr   typx.MultiRange[float64]
	}{
		{
			name: "empty",
			mr:   typx.MultiRange[float64]{},
		},
		{
			name: "single closed-closed",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: 1.5}, Upper: typx.Bound[float64]{Val: 10.5}},
			},
		},
		{
			name: "two non-overlapping with mixed bounds",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: 1.0}, Upper: typx.Bound[float64]{Val: 5.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 7.0, Exclusive: true}, Upper: typx.Bound[float64]{Val: 10.0}},
			},
		},
		{
			name: "three ranges",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: -20.0}, Upper: typx.Bound[float64]{Val: -10.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 0.0}, Upper: typx.Bound[float64]{Val: 5.5}},
				{Lower: typx.Bound[float64]{Val: 100.0, Exclusive: true}, Upper: typx.Bound[float64]{Val: 200.0}},
			},
		},
		{
			name: "unbounded lower bound",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Unbounded: true}, Upper: typx.Bound[float64]{Val: 5.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 7.0}, Upper: typx.Bound[float64]{Val: 10.0}},
			},
		},
		{
			name: "unbounded upper bound",
			mr: typx.MultiRange[float64]{
				{Lower: typx.Bound[float64]{Val: -5.0}, Upper: typx.Bound[float64]{Val: 0.0, Exclusive: true}},
				{Lower: typx.Bound[float64]{Val: 3.0}, Upper: typx.Bound[float64]{Unbounded: true}},
			},
		},
	}

	type doc struct {
		Val typx.MultiRange[float64] `bson:"val"`
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
