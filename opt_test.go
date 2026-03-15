package typx_test

import (
	"encoding/json"
	"testing"

	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type optField struct {
	Field typx.Opt[string] `json:"field,omitzero" bson:"field,omitempty"`
}

type optNilField struct {
	Field typx.Opt[typx.Nil[string]] `json:"field,omitzero" bson:"field,omitempty"`
}

func Test_Opt_JSON_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    []byte
		wantErr bool
	}{
		{
			name:  "set string",
			value: typx.OptFrom("example"),
			want:  []byte(`"example"`),
		},
		{
			name:  "set int",
			value: typx.OptFrom(42),
			want:  []byte(`42`),
		},
		{
			name:  "set object",
			value: typx.OptFrom(struct{ Name string }{Name: "test"}),
			want:  []byte(`{"Name":"test"}`),
		},
		{
			name:  "set Nil value",
			value: typx.OptFrom(typx.NilFrom("example")),
			want:  []byte(`"example"`),
		},
		{
			name:  "set Nil null",
			value: typx.OptFrom(typx.Nil[string]{}),
			want:  []byte(`null`),
		},
		{
			name:    "unset",
			value:   typx.Opt[string]{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Opt_JSON_Unmarshal(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  any
	}{
		{
			name:  "string",
			value: []byte(`"example"`),
			want:  typx.OptFrom("example"),
		},
		{
			name:  "int",
			value: []byte(`42`),
			want:  typx.OptFrom(42),
		},
		{
			name:  "object",
			value: []byte(`{"Name":"test"}`),
			want:  typx.OptFrom(struct{ Name string }{Name: "test"}),
		},
		{
			name:  "Nil value",
			value: []byte(`"example"`),
			want:  typx.OptFrom(typx.NilFrom("example")),
		},
		{
			name:  "Nil null",
			value: []byte(`null`),
			want:  typx.OptFrom(typx.Nil[string]{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case typx.Opt[string]:
				got := typx.Opt[string]{}
				err := json.Unmarshal(tt.value, &got)
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			case typx.Opt[int]:
				got := typx.Opt[int]{}
				err := json.Unmarshal(tt.value, &got)
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			case typx.Opt[struct{ Name string }]:
				got := typx.Opt[struct{ Name string }]{}
				err := json.Unmarshal(tt.value, &got)
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			case typx.Opt[typx.Nil[string]]:
				got := typx.Opt[typx.Nil[string]]{}
				err := json.Unmarshal(tt.value, &got)
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Opt_JSON_Struct_Omitzero(t *testing.T) {
	tests := []struct {
		name  string
		value optField
		want  []byte
	}{
		{
			name:  "field absent",
			value: optField{},
			want:  []byte(`{}`),
		},
		{
			name:  "field set to empty string",
			value: optField{Field: typx.OptFrom("")},
			want:  []byte(`{"field":""}`),
		},
		{
			name:  "field set to value",
			value: optField{Field: typx.OptFrom("example")},
			want:  []byte(`{"field":"example"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Opt_JSON_Struct_Omitzero_Nil(t *testing.T) {
	tests := []struct {
		name  string
		value optNilField
		want  []byte
	}{
		{
			name:  "field absent",
			value: optNilField{},
			want:  []byte(`{}`),
		},
		{
			name:  "field set to null",
			value: optNilField{Field: typx.OptFrom(typx.Nil[string]{})},
			want:  []byte(`{"field":null}`),
		},
		{
			name:  "field set to value",
			value: optNilField{Field: typx.OptFrom(typx.NilFrom("example"))},
			want:  []byte(`{"field":"example"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Opt_JSON_Struct_Unmarshal_Omitzero(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  optField
	}{
		{
			name:  "field absent",
			value: []byte(`{}`),
			want:  optField{},
		},
		{
			name:  "field present",
			value: []byte(`{"field":"example"}`),
			want:  optField{Field: typx.OptFrom("example")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optField{}
			err := json.Unmarshal(tt.value, &got)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Opt_JSON_Struct_Unmarshal_Omitzero_Nil(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  optNilField
	}{
		{
			name:  "field absent",
			value: []byte(`{}`),
			want:  optNilField{},
		},
		{
			name:  "field null",
			value: []byte(`{"field":null}`),
			want:  optNilField{Field: typx.OptFrom(typx.Nil[string]{})},
		},
		{
			name:  "field present",
			value: []byte(`{"field":"example"}`),
			want:  optNilField{Field: typx.OptFrom(typx.NilFrom("example"))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optNilField{}
			err := json.Unmarshal(tt.value, &got)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Opt_BSON_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    []byte
		wantErr bool
	}{
		{
			name:  "set string",
			value: typx.OptFrom("example"),
			want:  []byte("\b\x00\x00\x00example\x00"),
		},
		{
			name:  "set int",
			value: typx.OptFrom(42),
			want:  []byte{'*', 0x00, 0x00, 0x00},
		},
		{
			name:    "unset",
			value:   typx.Opt[string]{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := bson.MarshalValue(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_Opt_BSON_Unmarshal(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  any
	}{
		{
			name:  "string",
			value: []byte("\b\x00\x00\x00example\x00"),
			want:  typx.OptFrom("example"),
		},
		{
			name:  "int",
			value: []byte{'*', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			want:  typx.OptFrom(int64(42)),
		},
		{
			name:  "Nil null",
			value: []byte{},
			want:  typx.OptFrom(typx.Nil[string]{}),
		},
		{
			name:  "Nil value",
			value: []byte("\b\x00\x00\x00example\x00"),
			want:  typx.OptFrom(typx.NilFrom("example")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want.(type) {
			case typx.Opt[string]:
				got := typx.Opt[string]{}
				err := bson.UnmarshalValue(bson.TypeString, tt.value, &got)
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			case typx.Opt[int64]:
				got := typx.Opt[int64]{}
				err := bson.UnmarshalValue(bson.TypeInt64, tt.value, &got)
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			case typx.Opt[typx.Nil[string]]:
				switch tt.name {
				case "Nil null":
					got := typx.Opt[typx.Nil[string]]{}
					err := bson.UnmarshalValue(bson.TypeNull, tt.value, &got)
					require.NoError(t, err)
					require.Equal(t, tt.want, got)
				case "Nil value":
					got := typx.Opt[typx.Nil[string]]{}
					err := bson.UnmarshalValue(bson.TypeString, tt.value, &got)
					require.NoError(t, err)
					require.Equal(t, tt.want, got)
				}
			}
		})
	}
}

func Test_Opt_BSON_Struct_Omitempty(t *testing.T) {
	emptyDoc, _ := bson.Marshal(bson.D{})
	withEmpty, _ := bson.Marshal(bson.D{{Key: "field", Value: ""}})
	withValue, _ := bson.Marshal(bson.D{{Key: "field", Value: "example"}})
	tests := []struct {
		name  string
		value optField
		want  []byte
	}{
		{
			name:  "field absent",
			value: optField{},
			want:  emptyDoc,
		},
		{
			name:  "field set to empty string",
			value: optField{Field: typx.OptFrom("")},
			want:  withEmpty,
		},
		{
			name:  "field set to value",
			value: optField{Field: typx.OptFrom("example")},
			want:  withValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bson.Marshal(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Opt_BSON_Struct_Omitempty_Nil(t *testing.T) {
	emptyDoc, _ := bson.Marshal(bson.D{})
	withNull, _ := bson.Marshal(bson.D{{Key: "field", Value: nil}})
	withValue, _ := bson.Marshal(bson.D{{Key: "field", Value: "example"}})
	tests := []struct {
		name  string
		value optNilField
		want  []byte
	}{
		{
			name:  "field absent",
			value: optNilField{},
			want:  emptyDoc,
		},
		{
			name:  "field set to null",
			value: optNilField{Field: typx.OptFrom(typx.Nil[string]{})},
			want:  withNull,
		},
		{
			name:  "field set to value",
			value: optNilField{Field: typx.OptFrom(typx.NilFrom("example"))},
			want:  withValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bson.Marshal(tt.value)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Opt_BSON_Struct_Unmarshal_Omitempty(t *testing.T) {
	absentDoc, _ := bson.Marshal(bson.D{})
	presentDoc, _ := bson.Marshal(bson.D{{Key: "field", Value: "example"}})
	tests := []struct {
		name  string
		value []byte
		want  optField
	}{
		{
			name:  "field absent",
			value: absentDoc,
			want:  optField{},
		},
		{
			name:  "field present",
			value: presentDoc,
			want:  optField{Field: typx.OptFrom("example")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optField{}
			err := bson.Unmarshal(tt.value, &got)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Opt_BSON_Struct_Unmarshal_Omitempty_Nil(t *testing.T) {
	absentDoc, _ := bson.Marshal(bson.D{})
	nullDoc, _ := bson.Marshal(bson.D{{Key: "field", Value: nil}})
	presentDoc, _ := bson.Marshal(bson.D{{Key: "field", Value: "example"}})
	tests := []struct {
		name  string
		value []byte
		want  optNilField
	}{
		{
			name:  "field absent",
			value: absentDoc,
			want:  optNilField{},
		},
		{
			name:  "field null",
			value: nullDoc,
			want:  optNilField{Field: typx.OptFrom(typx.Nil[string]{})},
		},
		{
			name:  "field present",
			value: presentDoc,
			want:  optNilField{Field: typx.OptFrom(typx.NilFrom("example"))},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := optNilField{}
			err := bson.Unmarshal(tt.value, &got)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
