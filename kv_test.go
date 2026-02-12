package typx_test

import (
	"testing"

	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/assert"
)

func Test_KVFrom(t *testing.T) {
	k, v := "key", 1
	kv := typx.KVFrom(k, v)
	assert.Equal(t, k, kv.Key)
	assert.Equal(t, v, kv.Val)
}

func Test_KVsFrom(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	assert.Len(t, kvs, 2)
	assert.Equal(t, kv1, kvs[0])
	assert.Equal(t, kv2, kvs[1])
}

func Test_KVsFromMap(t *testing.T) {
	m := map[string]int{"k1": 1, "k2": 2}
	kvs := typx.KVsFromMap(m)

	assert.Len(t, kvs, 2)
	
	// Since map iteration order is random, we check if the elements exist
	dict := make(map[string]int)
	for _, kv := range kvs {
		dict[kv.Key] = kv.Val
	}
	assert.Equal(t, m, dict)
}

func Test_KVs_Keys(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	keys := kvs.Keys()
	assert.ElementsMatch(t, []string{"k1", "k2"}, keys)
}

func Test_KVs_Vals(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	vals := kvs.Vals()
	assert.ElementsMatch(t, []int{1, 2}, vals)
}

func Test_KVs_Map(t *testing.T) {
	kv1 := typx.KVFrom("k1", 1)
	kv2 := typx.KVFrom("k2", 2)
	kvs := typx.KVsFrom(kv1, kv2)

	m := kvs.Map()
	assert.Equal(t, map[string]int{"k1": 1, "k2": 2}, m)
}
