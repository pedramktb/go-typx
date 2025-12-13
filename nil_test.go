package typx_test

import (
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_Nil_Scan(t *testing.T) {
	randomID := uuid.New()
	tests := []struct {
		name  string
		value any
		want  any
	}{
		{
			name:  "null",
			value: nil,
			want:  typx.NilFromPtr[any](nil),
		},
		{
			name:  "string",
			value: "example",
			want:  typx.NilFrom("example"),
		},
		{
			name:  "object",
			value: randomID[:],
			want:  typx.NilFrom(randomID),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case typx.Nil[any]:
				got := typx.Nil[any]{}
				err := got.Scan(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[string]:
				got := typx.Nil[string]{}
				err := got.Scan(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[uuid.UUID]:
				got := typx.Nil[uuid.UUID]{}
				err := got.Scan(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Nil_Value(t *testing.T) {
	randomID := uuid.New()
	tests := []struct {
		name  string
		value any
		want  driver.Value
	}{
		{
			name:  "nil",
			value: typx.NilFromPtr[any](nil),
			want:  nil,
		},
		{
			name:  "string",
			value: typx.NilFrom("example"),
			want:  "example",
		},
		{
			name:  "uuid",
			value: typx.NilFrom(randomID),
			want:  randomID.String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.(driver.Valuer).Value()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Nil_Binary_Marshal(t *testing.T) {
	randomID := uuid.New()
	randomBinary, err := randomID.MarshalBinary()
	assert.NoError(t, err)

	tests := []struct {
		name  string
		value any
		want  []byte
	}{
		{
			name:  "string",
			value: typx.NilFrom("example"),
			want:  []byte("example"),
		},
		{
			name:  "null",
			value: typx.NilFromPtr[any](nil),
			want:  []byte(nil),
		},
		{
			name:  "uuid",
			value: randomID,
			want:  randomBinary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.(encoding.BinaryMarshaler).MarshalBinary()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Nil_Binary_Unmarshal(t *testing.T) {
	randomID := uuid.New()
	randomBinary, err := randomID.MarshalBinary()
	assert.NoError(t, err)

	tests := []struct {
		name  string
		value []byte
		want  any
	}{
		{
			name:  "string",
			value: []byte("example"),
			want:  typx.NilFrom("example"),
		},
		{
			name:  "null",
			value: nil,
			want:  typx.NilFromPtr[any](nil),
		},
		{
			name:  "uuid",
			value: randomBinary,
			want:  randomID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case typx.Nil[any]:
				got := typx.Nil[any]{}
				err := got.UnmarshalBinary(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[string]:
				got := typx.Nil[string]{}
				err := got.UnmarshalBinary(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[uuid.UUID]:
				got := typx.Nil[uuid.UUID]{}
				err := got.UnmarshalBinary(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Nil_Text_Marshal(t *testing.T) {
	randomID := uuid.New()
	randomText, err := randomID.MarshalText()
	assert.NoError(t, err)

	tests := []struct {
		name  string
		value any
		want  []byte
	}{
		{
			name:  "string",
			value: typx.NilFrom("example"),
			want:  []byte("example"),
		},
		{
			name:  "null",
			value: typx.NilFromPtr[any](nil),
			want:  []byte("null"),
		},
		{
			name:  "uuid",
			value: randomID,
			want:  randomText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.(encoding.TextMarshaler).MarshalText()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Nil_Text_Unmarshal(t *testing.T) {
	randomID := uuid.New()
	randomText, err := randomID.MarshalText()
	assert.NoError(t, err)

	tests := []struct {
		name  string
		value []byte
		want  any
	}{
		{
			name:  "string",
			value: []byte("example"),
			want:  typx.NilFrom("example"),
		},
		{
			name:  "null",
			value: nil,
			want:  typx.NilFromPtr[any](nil),
		},
		{
			name:  "uuid",
			value: randomText,
			want:  randomID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case typx.Nil[any]:
				got := typx.Nil[any]{}
				err := got.UnmarshalText(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[string]:
				got := typx.Nil[string]{}
				err := got.UnmarshalText(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[uuid.UUID]:
				got := typx.Nil[uuid.UUID]{}
				err := got.UnmarshalText(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

type field struct {
	Field typx.Nil[string] `json:"field"`
}

func Test_Nil_JSON_Marshal(t *testing.T) {
	randomID := uuid.New()

	tests := []struct {
		name  string
		value any
		want  []byte
	}{
		{
			name:  "string",
			value: typx.NilFrom("example"),
			want:  []byte(`"example"`),
		},
		{
			name:  "null",
			value: typx.NilFromPtr[any](nil),
			want:  []byte("null"),
		},
		{
			name:  "object",
			value: typx.NilFrom(struct{ ID uuid.UUID }{ID: randomID}),
			want:  []byte(`{"ID":"` + randomID.String() + `"}`),
		},
		{
			name: "object with null field",
			value: field{
				Field: typx.Nil[string]{},
			},
			want: []byte(`{"field":null}`),
		},
		{
			name: "object with valid field",
			value: field{
				Field: typx.Nil[string]{Val: "example", NotNil: true},
			},
			want: []byte(`{"field":"example"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Nil_JSON_UnMarshal(t *testing.T) {
	randomID := uuid.New()
	tests := []struct {
		name  string
		value []byte
		want  any
	}{
		{
			name:  "nil",
			value: []byte("null"),
			want:  typx.NilFromPtr[any](nil),
		},
		{
			name:  "string",
			value: []byte(`"example"`),
			want:  typx.NilFrom("example"),
		},
		{
			name:  "object",
			value: []byte(`{"ID":"` + randomID.String() + `"}`),
			want:  typx.NilFrom(struct{ ID uuid.UUID }{ID: randomID}),
		},
		{
			name:  "object with null field",
			value: []byte(`{"field":null}`),
			want:  field{Field: typx.Nil[string]{}},
		},
		{
			name:  "object with undefined field",
			value: []byte(`{}`),
			want:  field{Field: typx.Nil[string]{}},
		},
		{
			name:  "object with valid field",
			value: []byte(`{"field":"example"}`),
			want: field{
				Field: typx.Nil[string]{Val: "example", NotNil: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case typx.Nil[any]:
				got := typx.Nil[any]{}
				err := json.Unmarshal(tt.value, &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[string]:
				got := typx.Nil[string]{}
				err := json.Unmarshal(tt.value, &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[struct{ ID uuid.UUID }]:
				got := typx.Nil[struct{ ID uuid.UUID }]{}
				err := json.Unmarshal(tt.value, &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case field:
				got := field{}
				err := json.Unmarshal(tt.value, &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Nil_BSON_Marshal(t *testing.T) {
	randomID := uuid.New()
	randomBSON, _ := bson.Marshal(struct{ ID uuid.UUID }{ID: randomID})
	tests := []struct {
		name  string
		value any
		want  []byte
	}{
		{
			name:  "string",
			value: typx.NilFrom("example"),
			want:  []byte("\b\x00\x00\x00example\x00"),
		},
		{
			name:  "null",
			value: typx.NilFromPtr[any](nil),
			want:  []byte{},
		},
		{
			name:  "object",
			value: typx.NilFrom(struct{ ID uuid.UUID }{ID: randomID}),
			want:  randomBSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := bson.MarshalValue(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Nil_BSON_UnMarshal(t *testing.T) {
	randomID := uuid.New()
	randomBSON, _ := bson.Marshal(struct{ ID uuid.UUID }{ID: randomID})
	tests := []struct {
		name  string
		value []byte
		want  any
	}{
		{
			name:  "null",
			value: []byte{},
			want:  typx.NilFromPtr[any](nil),
		},
		{
			name:  "string",
			value: []byte("\b\x00\x00\x00example\x00"),
			want:  typx.NilFrom("example"),
		},
		{
			name:  "object",
			value: randomBSON,
			want:  typx.NilFrom(struct{ ID uuid.UUID }{ID: randomID}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case typx.Nil[any]:
				got := typx.Nil[any]{}
				err := bson.UnmarshalValue(bson.TypeNull, tt.value, &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[string]:
				got := typx.Nil[string]{}
				err := bson.UnmarshalValue(bson.TypeString, tt.value, &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			case typx.Nil[struct{ ID uuid.UUID }]:
				got := typx.Nil[struct{ ID uuid.UUID }]{}
				err := bson.UnmarshalValue(bson.TypeEmbeddedDocument, tt.value, &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
