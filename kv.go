package typx

type KV[K comparable, V any] struct {
	Key K
	Val V
}

// KVFrom creates a KV pair from the given key and value.
func KVFrom[K comparable, V any](key K, val V) KV[K, V] {
	return KV[K, V]{Key: key, Val: val}
}

type KVs[K comparable, V any] []KV[K, V]

func KVsFrom[K comparable, V any](kvs ...KV[K, V]) KVs[K, V] {
	return kvs
}

func KVsFromMap[K comparable, V any](m map[K]V) KVs[K, V] {
	kvs := make(KVs[K, V], 0, len(m))
	for k, v := range m {
		kvs = append(kvs, KVFrom(k, v))
	}
	return kvs
}

func (kvs KVs[K, V]) Keys() []K {
	keys := make([]K, len(kvs))
	for i, kv := range kvs {
		keys[i] = kv.Key
	}
	return keys
}

func (kvs KVs[K, V]) Vals() []V {
	vals := make([]V, len(kvs))
	for i, kv := range kvs {
		vals[i] = kv.Val
	}
	return vals
}

func (kvs KVs[K, V]) Map() map[K]V {
	m := make(map[K]V, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Val
	}
	return m
}
