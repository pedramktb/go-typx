package typx

// KV represents a key-value pair with a comparable key and any value.
type KV[K comparable, V any] struct {
	Key K
	Val V
}

// KVFrom creates a KV pair from the given key and value.
func KVFrom[K comparable, V any](key K, val V) KV[K, V] {
	return KV[K, V]{Key: key, Val: val}
}

// KVs is a slice of KV pairs that provides utility methods for working with key-value collections.
type KVs[K comparable, V any] []KV[K, V]

// KVsFrom creates a KVs slice from the given KV pairs.
func KVsFrom[K comparable, V any](kvs ...KV[K, V]) KVs[K, V] {
	return kvs
}

// KVsFromMap creates a KVs slice from the given map of key-value pairs.
func KVsFromMap[K comparable, V any](m map[K]V) KVs[K, V] {
	kvs := make(KVs[K, V], 0, len(m))
	for k, v := range m {
		kvs = append(kvs, KVFrom(k, v))
	}
	return kvs
}

// Keys returns a slice of all keys in the KVs collection.
func (kvs KVs[K, V]) Keys() []K {
	keys := make([]K, len(kvs))
	for i, kv := range kvs {
		keys[i] = kv.Key
	}
	return keys
}

// Vals returns a slice of all values in the KVs collection.
func (kvs KVs[K, V]) Vals() []V {
	vals := make([]V, len(kvs))
	for i, kv := range kvs {
		vals[i] = kv.Val
	}
	return vals
}

// Map converts the KVs collection into a map of key-value pairs.
func (kvs KVs[K, V]) Map() map[K]V {
	m := make(map[K]V, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Val
	}
	return m
}
