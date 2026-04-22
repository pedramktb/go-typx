package typx

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// [KV] represents a key-value pair with a comparable key and any value.
type KV[K comparable, V any] struct {
	Key K
	Val V
}

// [KVFrom] creates a [KV] pair from the given key and value.
func KVFrom[K comparable, V any](key K, val V) KV[K, V] {
	return KV[K, V]{Key: key, Val: val}
}

// [KVs] is a slice of [KV] pairs that provides utility methods for working with key-value collections.
type KVs[K comparable, V any] []KV[K, V]

// [KVsFrom] creates a [KVs] slice from the given [KV] pairs.
func KVsFrom[K comparable, V any](kvs ...KV[K, V]) KVs[K, V] {
	return kvs
}

// [KVsFromMap] creates a [KVs] slice from the given map of key-value pairs.
func KVsFromMap[K comparable, V any](m map[K]V) KVs[K, V] {
	kvs := make(KVs[K, V], 0, len(m))
	for k, v := range m {
		kvs = append(kvs, KV[K, V]{Key: k, Val: v})
	}
	return kvs
}

// Keys returns a slice of all keys in the [KVs] collection.
func (kvs KVs[K, V]) Keys() []K {
	keys := make([]K, len(kvs))
	for i, kv := range kvs {
		keys[i] = kv.Key
	}
	return keys
}

// Vals returns a slice of all values in the [KVs] collection.
func (kvs KVs[K, V]) Vals() []V {
	vals := make([]V, len(kvs))
	for i, kv := range kvs {
		vals[i] = kv.Val
	}
	return vals
}

// Map converts the [KVs] collection into a map of key-value pairs.
func (kvs KVs[K, V]) Map() map[K]V {
	m := make(map[K]V, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Val
	}
	return m
}

var _ json.Marshaler = KVs[any, any]{}

// MarshalJSON implements the [json.Marshaler] interface.
// [KVs] is serialized as a single JSON object: {"k1": v1, "k2": v2, ...}.
func (kvs KVs[K, V]) MarshalJSON() ([]byte, error) {
	m := make(map[K]V, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Val
	}
	return json.Marshal(m)
}

var _ json.Unmarshaler = (*KVs[any, any])(nil)

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (kvs *KVs[K, V]) UnmarshalJSON(data []byte) error {
	var m map[K]V
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*kvs = make(KVs[K, V], 0, len(m))
	for k, v := range m {
		*kvs = append(*kvs, KV[K, V]{Key: k, Val: v})
	}
	return nil
}

var _ driver.Valuer = (*KVs[any, any])(nil)

// Scan implements the [sql.Scanner] interface.
// The column should be a type that can hold JSON data (JSONB, JSON, TEXT, etc).
func (kvs *KVs[K, V]) Scan(src any) error {
	if src == nil {
		*kvs = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, kvs)
	case string:
		return json.Unmarshal([]byte(v), kvs)
	}
	return fmt.Errorf("typx.KVs.Scan: %T is not a string or []byte", src)
}

var _ driver.Valuer = (*KVs[any, any])(nil)

// Value implements the [driver.Valuer] interface.
func (kvs KVs[K, V]) Value() (driver.Value, error) {
	return json.Marshal(kvs)
}

var _ bson.ValueMarshaler = (*KVs[any, any])(nil)

// MarshalBSONValue implements the [bson.ValueMarshaler] interface.
// [KVs] is serialized as a single BSON document: {k1: v1, k2: v2, ...}.
func (kvs KVs[K, V]) MarshalBSONValue() (byte, []byte, error) {
	m := make(map[K]V, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Val
	}
	t, data, err := bson.MarshalValue(m)
	return byte(t), data, err
}

var _ bson.ValueUnmarshaler = (*KVs[any, any])(nil)

// UnmarshalBSONValue implements the [bson.ValueUnmarshaler] interface.
func (kvs *KVs[K, V]) UnmarshalBSONValue(t byte, data []byte) error {
	var m map[K]V
	if err := bson.UnmarshalValue(bson.Type(t), data, &m); err != nil {
		return err
	}
	*kvs = make(KVs[K, V], 0, len(m))
	for k, v := range m {
		*kvs = append(*kvs, KV[K, V]{Key: k, Val: v})
	}
	return nil
}

var _ bson.ValueMarshaler = (*KVs[any, any])(nil)

// MarshalBSON implements the [bson.Marshaler] interface for use as a standalone document.
func (kvs KVs[K, V]) MarshalBSON() ([]byte, error) {
	m := make(map[K]V, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Val
	}
	return bson.Marshal(m)
}

var _ bson.Marshaler = (*KVs[any, any])(nil)

// UnmarshalBSON implements the [bson.Unmarshaler] interface for use as a standalone document.
func (kvs *KVs[K, V]) UnmarshalBSON(data []byte) error {
	var m map[K]V
	if err := bson.Unmarshal(data, &m); err != nil {
		return err
	}
	*kvs = make(KVs[K, V], 0, len(m))
	for k, v := range m {
		*kvs = append(*kvs, KV[K, V]{Key: k, Val: v})
	}
	return nil
}
