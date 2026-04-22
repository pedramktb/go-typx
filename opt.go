package typx

import (
	"encoding/json"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// [Opt] is a type that can be used to represent an optional value (present vs. absent).
// It should be used for fields that may be missing entirely in serialized form.
// For nullable values (present but null), use [Nil] instead.
// Combining both is possible: [Opt][[Nil][T]] represents a field that can be absent, null, or a value.
//
// Use the `omitzero` JSON struct tag to omit absent fields. Example:
//
//	type Struct struct {
//		Field Opt[int] `json:"field,omitzero"`
//	}
type Opt[T any] struct {
	Val T
	Set bool
}

// OptFrom creates an [Opt] that is set to the given value.
func OptFrom[T any](value T) Opt[T] {
	return Opt[T]{Val: value, Set: true}
}

// OptFromPtr creates an [Opt] from a pointer. If the pointer is nil, the result is unset.
func OptFromPtr[T any](ptr *T) Opt[T] {
	if ptr == nil {
		return Opt[T]{Set: false}
	}
	return Opt[T]{Val: *ptr, Set: true}
}

// IsZero returns true when the value is absent.
// This enables the `omitzero` struct tag in encoding/json (Go 1.24+).
func (o Opt[T]) IsZero() bool { return !o.Set }

var _ json.Marshaler = Opt[any]{}

// MarshalJSON marshals the inner value transparently.
// Returns an error if the value is not set, to prevent silent null serialization.
// Use the `omitzero` struct tag to omit absent fields instead of marshaling them directly.
func (o Opt[T]) MarshalJSON() ([]byte, error) {
	if !o.Set {
		return nil, errors.New("typx.Opt.MarshalJSON: cannot marshal unset value; use omitzero struct tag to omit absent fields")
	}
	return json.Marshal(o.Val)
}

var _ json.Unmarshaler = (*Opt[any])(nil)

// UnmarshalJSON is called only when the field is present in the JSON input (even as null).
// It always marks the [Opt] as set and delegates to the inner type.
func (o *Opt[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	return json.Unmarshal(data, &o.Val)
}

var _ bson.ValueMarshaler = Opt[any]{}

// MarshalBSONValue implements the [bson.ValueMarshaler] interface.
// Returns an error if the value is not set, to prevent silent null serialization.
// Use the `omitempty` struct tag to omit absent fields instead of marshaling them directly.
func (o Opt[T]) MarshalBSONValue() (byte, []byte, error) {
	if !o.Set {
		return byte(bson.Type(0)), nil, errors.New("Opt.MarshalBSONValue: cannot marshal unset value; use omitempty struct tag to omit absent fields")
	}
	t, data, err := bson.MarshalValue(o.Val)
	return byte(t), data, err
}

var _ bson.ValueUnmarshaler = (*Opt[any])(nil)

// UnmarshalBSONValue implements the [bson.ValueUnmarshaler] interface.
// It always marks the Opt as set and delegates to the inner type.
func (o *Opt[T]) UnmarshalBSONValue(t byte, data []byte) error {
	o.Set = true
	return bson.UnmarshalValue(bson.Type(t), data, &o.Val)
}
