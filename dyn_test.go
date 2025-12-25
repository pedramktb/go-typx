package typx_test

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func Test_Dyn_Scan(t *testing.T) {
	randomID := uuid.New()
	tests := []struct {
		name    string
		value   any
		want    any
		wantErr bool
	}{
		{
			name:  "null",
			value: nil,
			want:  typx.Dyn{Val: nil},
		},
		{
			name:  "string json",
			value: []byte(`"example"`),
			want:  typx.Dyn{Val: "example"},
		},
		{
			name:  "string type",
			value: `"example"`,
			want:  typx.Dyn{Val: "example"},
		},
		{
			name:  "number json",
			value: []byte(`42`),
			want:  typx.Dyn{Val: float64(42)},
		},
		{
			name:  "object json",
			value: []byte(`{"id":"` + randomID.String() + `"}`),
			want:  typx.Dyn{Val: map[string]any{"id": randomID.String()}},
		},
		{
			name:  "array json",
			value: []byte(`[1,2,3]`),
			want:  typx.Dyn{Val: []any{float64(1), float64(2), float64(3)}},
		},
		{
			name:    "invalid type",
			value:   123,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typx.Dyn{}
			err := got.Scan(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Dyn_Value(t *testing.T) {
	randomID := uuid.New()
	tests := []struct {
		name  string
		value typx.Dyn
		want  driver.Value
	}{
		{
			name:  "nil",
			value: typx.Dyn{Val: nil},
			want:  []byte("null"),
		},
		{
			name:  "string",
			value: typx.Dyn{Val: "example"},
			want:  []byte(`"example"`),
		},
		{
			name:  "number",
			value: typx.Dyn{Val: 42},
			want:  []byte("42"),
		},
		{
			name:  "boolean",
			value: typx.Dyn{Val: true},
			want:  []byte("true"),
		},
		{
			name:  "object",
			value: typx.Dyn{Val: map[string]any{"id": randomID.String()}},
			want:  []byte(`{"id":"` + randomID.String() + `"}`),
		},
		{
			name:  "array",
			value: typx.Dyn{Val: []any{1, 2, 3}},
			want:  []byte("[1,2,3]"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.Value()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Dyn_Binary_Marshal(t *testing.T) {
	randomID := uuid.New()
	randomBinary, err := randomID.MarshalBinary()
	assert.NoError(t, err)

	tests := []struct {
		name    string
		value   typx.Dyn
		want    []byte
		wantErr bool
	}{
		{
			name:  "uuid",
			value: typx.Dyn{Val: randomID},
			want:  randomBinary,
		},
		{
			name:    "string without interface",
			value:   typx.Dyn{Val: "example"},
			wantErr: true,
		},
		{
			name:    "nil",
			value:   typx.Dyn{Val: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.MarshalBinary()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Dyn_Binary_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		initial any
		value   []byte
		want    typx.Dyn
		wantErr bool
	}{
		{
			name:    "string without interface",
			initial: "example",
			value:   []byte("example"),
			wantErr: true,
		},
		{
			name:    "nil",
			initial: nil,
			value:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typx.Dyn{Val: tt.initial}
			err := got.UnmarshalBinary(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Dyn_Text_Marshal(t *testing.T) {
	randomID := uuid.New()
	randomText, err := randomID.MarshalText()
	assert.NoError(t, err)

	tests := []struct {
		name    string
		value   typx.Dyn
		want    []byte
		wantErr bool
	}{
		{
			name:  "uuid",
			value: typx.Dyn{Val: randomID},
			want:  randomText,
		},
		{
			name:    "string without interface",
			value:   typx.Dyn{Val: "example"},
			wantErr: true,
		},
		{
			name:    "nil",
			value:   typx.Dyn{Val: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.MarshalText()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Dyn_Text_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		initial any
		value   []byte
		want    typx.Dyn
		wantErr bool
	}{
		{
			name:    "string without interface",
			initial: "example",
			value:   []byte("example"),
			wantErr: true,
		},
		{
			name:    "nil",
			initial: nil,
			value:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typx.Dyn{Val: tt.initial}
			err := got.UnmarshalText(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

type dynField struct {
	Field typx.Dyn `json:"field"`
}

func Test_Dyn_JSON_Marshal(t *testing.T) {
	randomID := uuid.New()

	tests := []struct {
		name  string
		value any
		want  []byte
	}{
		{
			name:  "string",
			value: typx.Dyn{Val: "example"},
			want:  []byte(`"example"`),
		},
		{
			name:  "null",
			value: typx.Dyn{Val: nil},
			want:  []byte("null"),
		},
		{
			name:  "number",
			value: typx.Dyn{Val: 42},
			want:  []byte("42"),
		},
		{
			name:  "boolean",
			value: typx.Dyn{Val: true},
			want:  []byte("true"),
		},
		{
			name:  "object",
			value: typx.Dyn{Val: map[string]any{"ID": randomID.String()}},
			want:  []byte(`{"ID":"` + randomID.String() + `"}`),
		},
		{
			name:  "array",
			value: typx.Dyn{Val: []any{1, 2, 3}},
			want:  []byte(`[1,2,3]`),
		},
		{
			name: "object with null field",
			value: dynField{
				Field: typx.Dyn{Val: nil},
			},
			want: []byte(`{"field":null}`),
		},
		{
			name: "object with valid field",
			value: dynField{
				Field: typx.Dyn{Val: "example"},
			},
			want: []byte(`{"field":"example"}`),
		},
		{
			name: "nested object",
			value: dynField{
				Field: typx.Dyn{Val: map[string]any{"nested": "value"}},
			},
			want: []byte(`{"field":{"nested":"value"}}`),
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

func Test_Dyn_JSON_UnMarshal(t *testing.T) {
	randomID := uuid.New()
	tests := []struct {
		name  string
		value []byte
		want  typx.Dyn
	}{
		{
			name:  "nil",
			value: []byte("null"),
			want:  typx.Dyn{Val: nil},
		},
		{
			name:  "string",
			value: []byte(`"example"`),
			want:  typx.Dyn{Val: "example"},
		},
		{
			name:  "number",
			value: []byte(`42`),
			want:  typx.Dyn{Val: float64(42)},
		},
		{
			name:  "boolean",
			value: []byte(`true`),
			want:  typx.Dyn{Val: true},
		},
		{
			name:  "object",
			value: []byte(`{"ID":"` + randomID.String() + `"}`),
			want:  typx.Dyn{Val: map[string]any{"ID": randomID.String()}},
		},
		{
			name:  "array",
			value: []byte(`[1,2,3]`),
			want:  typx.Dyn{Val: []any{float64(1), float64(2), float64(3)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typx.Dyn{}
			err := json.Unmarshal(tt.value, &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Dyn_JSON_UnMarshal_InStruct(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  dynField
	}{
		{
			name:  "object with null field",
			value: []byte(`{"field":null}`),
			want:  dynField{Field: typx.Dyn{Val: nil}},
		},
		{
			name:  "object with undefined field",
			value: []byte(`{}`),
			want:  dynField{Field: typx.Dyn{Val: nil}},
		},
		{
			name:  "object with string field",
			value: []byte(`{"field":"example"}`),
			want:  dynField{Field: typx.Dyn{Val: "example"}},
		},
		{
			name:  "object with number field",
			value: []byte(`{"field":42}`),
			want:  dynField{Field: typx.Dyn{Val: float64(42)}},
		},
		{
			name:  "object with nested object",
			value: []byte(`{"field":{"nested":"value"}}`),
			want:  dynField{Field: typx.Dyn{Val: map[string]any{"nested": "value"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dynField{}
			err := json.Unmarshal(tt.value, &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Dyn_BSON_Marshal(t *testing.T) {
	randomID := uuid.New()
	objBSON, _ := bson.Marshal(map[string]any{"ID": randomID.String()})
	tests := []struct {
		name    string
		value   typx.Dyn
		want    []byte
		wantErr bool
	}{
		{
			name:  "string",
			value: typx.Dyn{Val: "example"},
			want:  []byte("\b\x00\x00\x00example\x00"),
		},

		{
			name:    "null",
			value:   typx.Dyn{Val: nil},
			wantErr: true,
		},
		{
			name:  "number",
			value: typx.Dyn{Val: int32(42)},
			want:  []byte("*\x00\x00\x00"),
		},
		{
			name:  "object",
			value: typx.Dyn{Val: map[string]any{"ID": randomID.String()}},
			want:  objBSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := tt.value.MarshalBSONValue()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Dyn_BSON_UnMarshal(t *testing.T) {
	randomID := uuid.New()
	objBSON, _ := bson.Marshal(map[string]any{"ID": randomID.String()})

	nestedObj := map[string]any{
		"name": "test",
		"nested": map[string]any{
			"value": 42,
			"items": []any{"a", "b", "c"},
		},
	}
	nestedBSON, _ := bson.Marshal(nestedObj)

	tests := []struct {
		name     string
		bsonType bsontype.Type
		value    []byte
		want     typx.Dyn
	}{
		{
			name:     "null",
			bsonType: bson.TypeNull,
			value:    []byte{},
			want:     typx.Dyn{Val: nil},
		},
		{
			name:     "string",
			bsonType: bson.TypeString,
			value:    []byte("\b\x00\x00\x00example\x00"),
			want:     typx.Dyn{Val: "example"},
		},
		{
			name:     "number",
			bsonType: bson.TypeInt32,
			value:    []byte("*\x00\x00\x00"),
			want:     typx.Dyn{Val: int32(42)},
		},
		{
			name:     "object",
			bsonType: bson.TypeEmbeddedDocument,
			value:    objBSON,
			want:     typx.Dyn{Val: map[string]any{"ID": randomID.String()}},
		},
		{
			name:     "nested object with array",
			bsonType: bson.TypeEmbeddedDocument,
			value:    nestedBSON,
			want: typx.Dyn{Val: map[string]any{
				"name": "test",
				"nested": map[string]any{
					"value": int32(42),
					"items": []any{"a", "b", "c"},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typx.Dyn{}
			err := got.UnmarshalBSONValue(tt.bsonType, tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
