package typx_test

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func Test_KVFrom(t *testing.T) {
	k, v := "key", 1
	kv := typx.KVFrom(k, v)
	require.Equal(t, k, kv.Key)
	require.Equal(t, v, kv.Val)
}

func Test_KVsFrom(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	require.Len(t, kvs, 2)
	require.Equal(t, kv1, kvs[0])
	require.Equal(t, kv2, kvs[1])
}

func Test_KVsFromMap(t *testing.T) {
	m := map[string]int{"k1": 1, "k2": 2}
	kvs := typx.KVsFromMap(m)

	require.Len(t, kvs, 2)

	// Since map iteration order is random, we check if the elements exist
	dict := make(map[string]int)
	for _, kv := range kvs {
		dict[kv.Key] = kv.Val
	}
	require.Equal(t, m, dict)
}

func Test_KVs_Keys(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	keys := kvs.Keys()
	require.ElementsMatch(t, []string{"k1", "k2"}, keys)
}

func Test_KVs_Vals(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	vals := kvs.Vals()
	require.ElementsMatch(t, []int{1, 2}, vals)
}

func Test_KVs_Map(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	m := kvs.Map()
	require.Equal(t, map[string]int{"k1": 1, "k2": 2}, m)
}

func Test_KV_JSON_Marshal(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  []byte
	}{
		{
			name:  "string value",
			value: typx.KVFrom("mykey", "myval"),
			want:  []byte(`{"mykey":"myval"}`),
		},
		{
			name:  "int value",
			value: typx.KVFrom("count", 42),
			want:  []byte(`{"count":42}`),
		},
		{
			name:  "object value",
			value: typx.KVFrom("obj", struct{ Name string }{Name: "test"}),
			want:  []byte(`{"obj":{"Name":"test"}}`),
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

func Test_KV_JSON_Unmarshal(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want typx.KV[string, string]
	}{
		{
			name: "string value",
			data: []byte(`{"mykey":"myval"}`),
			want: typx.KVFrom("mykey", "myval"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got typx.KV[string, string]
			err := json.Unmarshal(tt.data, &got)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_KV_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    typx.KV[string, string]
		wantErr bool
	}{
		{
			name: "bytes",
			src:  []byte(`{"k":"v"}`),
			want: typx.KVFrom("k", "v"),
		},
		{
			name: "string",
			src:  `{"k":"v"}`,
			want: typx.KVFrom("k", "v"),
		},
		{
			name:    "invalid type",
			src:     123,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got typx.KV[string, string]
			err := got.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_KV_Value(t *testing.T) {
	tests := []struct {
		name  string
		value typx.KV[string, int]
		want  driver.Value
	}{
		{
			name:  "kv pair",
			value: typx.KVFrom("count", 42),
			want:  []byte(`{"count":42}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.Value()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_KV_BSON(t *testing.T) {
	type doc struct {
		KV typx.KV[string, int] `bson:"kv"`
	}
	original := doc{KV: typx.KVFrom("mykey", 7)}
	data, err := bson.Marshal(original)
	require.NoError(t, err)

	var got doc
	err = bson.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, original, got)
}

func Test_KV_BSON_Standalone(t *testing.T) {
	original := typx.KVFrom("mykey", 7)
	data, err := bson.Marshal(original)
	require.NoError(t, err)

	var got typx.KV[string, int]
	err = bson.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, original, got)
}

func Test_KVs_JSON_Marshal(t *testing.T) {
	kvs := typx.KVsFrom(typx.KVFrom("k1", 1), typx.KVFrom("k2", 2))
	got, err := json.Marshal(kvs)
	require.NoError(t, err)
	require.JSONEq(t, `{"k1":1,"k2":2}`, string(got))
}

func Test_KVs_JSON_Unmarshal(t *testing.T) {
	data := []byte(`{"k1":1,"k2":2}`)
	var got typx.KVs[string, int]
	err := json.Unmarshal(data, &got)
	require.NoError(t, err)
	require.ElementsMatch(t, typx.KVsFrom(typx.KVFrom("k1", 1), typx.KVFrom("k2", 2)), got)
}

func Test_KVs_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		want    typx.KVs[string, int]
		wantErr bool
	}{
		{
			name: "bytes",
			src:  []byte(`{"k1":1,"k2":2}`),
			want: typx.KVsFrom(typx.KVFrom("k1", 1), typx.KVFrom("k2", 2)),
		},
		{
			name: "string",
			src:  `{"k1":1}`,
			want: typx.KVsFrom(typx.KVFrom("k1", 1)),
		},
		{
			name: "null",
			src:  nil,
			want: nil,
		},
		{
			name:    "invalid type",
			src:     42,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got typx.KVs[string, int]
			err := got.Scan(tt.src)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tt.want, got)
			}
		})
	}
}

func Test_KVs_Value(t *testing.T) {
	kvs := typx.KVsFrom(typx.KVFrom("k1", 1), typx.KVFrom("k2", 2))
	got, err := kvs.Value()
	require.NoError(t, err)
	require.JSONEq(t, `{"k1":1,"k2":2}`, string(got.([]byte)))
}

func Test_KVs_BSON(t *testing.T) {
	type doc struct {
		KVs typx.KVs[string, int] `bson:"kvs"`
	}
	original := doc{KVs: typx.KVsFrom(typx.KVFrom("k1", 1), typx.KVFrom("k2", 2))}
	data, err := bson.Marshal(original)
	require.NoError(t, err)

	var got doc
	err = bson.Unmarshal(data, &got)
	require.NoError(t, err)
	require.ElementsMatch(t, original.KVs, got.KVs)
}

func Test_KVs_BSON_Standalone(t *testing.T) {
	original := typx.KVsFrom(typx.KVFrom("k1", 1), typx.KVFrom("k2", 2))
	data, err := bson.Marshal(original)
	require.NoError(t, err)

	var got typx.KVs[string, int]
	err = bson.Unmarshal(data, &got)
	require.NoError(t, err)
	require.ElementsMatch(t, original, got)
}
