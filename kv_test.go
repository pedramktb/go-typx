package typx_test

import (
	"encoding/json"
	"testing"

	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

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
