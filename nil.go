package typx

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// [Nil] is a type that can be used to represent a nil/nullable value.
// It implements most of the interfaces that are used to marshal and unmarshal values.
// For optional values that are not present, use [Opt] instead.
type Nil[T any] struct {
	Val    T
	NotNil bool
}

// NilFrom creates a [Nil] from a non-nil value.
func NilFrom[T any](value T) Nil[T] { return Nil[T]{Val: value, NotNil: true} }

// NilFromPtr creates a [Nil] from a pointer. If the pointer is nil, NotNil is false.
func NilFromPtr[T any](value *T) Nil[T] {
	if value == nil {
		return Nil[T]{}
	}
	return Nil[T]{Val: *value, NotNil: true}
}

// Ptr returns a pointer to the value if NotNil is true, otherwise nil.
// It uses a non-pointer receiver so that the modified pointer does not affect the original value.
func (n Nil[T]) Ptr() *T {
	if !n.NotNil {
		return nil
	}
	return &n.Val
}

var _ sql.Scanner = (*Nil[any])(nil)

// Scan implements the [sql.Scanner] interface.
func (n *Nil[T]) Scan(src any) error {
	n.NotNil = false
	if src == nil {
		n.Val = *new(T)
		return nil
	}
	if scanner, ok := any(&n.Val).(sql.Scanner); ok {
		err := scanner.Scan(src)
		if err != nil {
			return err
		}
		n.NotNil = true
		return nil
	}
	switch v := src.(type) {
	case T:
		n.Val = v
		n.NotNil = true
		return nil
	case []byte:
		if val, ok := any(string(v)).(T); ok {
			n.Val = val
			n.NotNil = true
			return nil
		}
	case string:
		if val, ok := any([]byte(v)).(T); ok {
			n.Val = val
			n.NotNil = true
			return nil
		}
	}
	return fmt.Errorf("typx.Nil.Scan: %T does not implement sql.Scanner and cannot be scanned from %T", n.Val, src)
}

var _ driver.Valuer = (*Nil[any])(nil)

// Value implements the [driver.Valuer] interface.
func (n Nil[T]) Value() (driver.Value, error) {
	if !n.NotNil {
		return nil, nil
	}
	if valuer, ok := any(n.Val).(driver.Valuer); ok {
		return valuer.Value()
	}
	return n.Val, nil
}

var _ encoding.BinaryMarshaler = Nil[any]{}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (n Nil[T]) MarshalBinary() ([]byte, error) {
	if !n.NotNil {
		return []byte(nil), nil
	}
	switch v := any(n.Val).(type) {
	case encoding.BinaryMarshaler:
		return v.MarshalBinary()
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	}
	return nil, fmt.Errorf("typx.Nil.MarshalBinary: %T does not implement encoding.BinaryMarshaler and is not a string or []byte", n.Val)
}

var _ encoding.BinaryUnmarshaler = (*Nil[any])(nil)

// UnmarshalBinary implements the [encoding.BinaryUnmarshaler] interface.
func (n *Nil[T]) UnmarshalBinary(data []byte) error {
	n.NotNil = false
	if data == nil {
		n.Val = *new(T)
		return nil
	}
	if unmarshaler, ok := any(n.Val).(encoding.BinaryUnmarshaler); ok {
		if err := unmarshaler.UnmarshalBinary(data); err != nil {
			return err
		}
		n.NotNil = true
		return nil
	}
	var ok bool
	if n.Val, ok = any(data).(T); ok {
		n.NotNil = true
		return nil
	}
	if n.Val, ok = any(string(data)).(T); ok {
		n.NotNil = true
		return nil
	}
	if n.Val, ok = any([]byte(data)).(T); ok {
		n.NotNil = true
		return nil
	}
	return fmt.Errorf("typx.Nil.UnmarshalBinary: %T does not implement encoding.BinaryUnmarshaler and is not a string or []byte", n.Val)
}

var _ encoding.TextMarshaler = Nil[any]{}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (n Nil[T]) MarshalText() ([]byte, error) {
	if !n.NotNil {
		return []byte("null"), nil
	}
	switch v := any(n.Val).(type) {
	case encoding.TextMarshaler:
		return v.MarshalText()
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	}
	return nil, fmt.Errorf("typx.Nil.MarshalText: %T does not implement encoding.TextMarshaler and is not a string or []byte", n.Val)
}

var _ encoding.TextUnmarshaler = (*Nil[any])(nil)

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (n *Nil[T]) UnmarshalText(data []byte) error {
	n.NotNil = false
	if data == nil {
		n.Val = *new(T)
		return nil
	}
	if unmarshaler, ok := any(n.Val).(encoding.TextUnmarshaler); ok {
		if err := unmarshaler.UnmarshalText(data); err != nil {
			return err
		}
		n.NotNil = true
		return nil
	}
	var ok bool
	if n.Val, ok = any(data).(T); ok {
		n.NotNil = true
		return nil
	}
	if n.Val, ok = any(string(data)).(T); ok {
		n.NotNil = true
		return nil
	}
	if n.Val, ok = any([]byte(data)).(T); ok {
		n.NotNil = true
		return nil
	}
	return fmt.Errorf("typx.Nil.UnmarshalText: %T does not implement encoding.TextUnmarshaler and is not a string or []byte", n.Val)
}

var _ json.Marshaler = Nil[any]{}

// MarshalJSON implements the [json.Marshaler] interface.
func (n Nil[T]) MarshalJSON() ([]byte, error) {
	if !n.NotNil {
		return []byte("null"), nil
	}
	return json.Marshal(n.Val)
}

var _ json.Unmarshaler = (*Nil[any])(nil)

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (n *Nil[T]) UnmarshalJSON(data []byte) error {
	n.NotNil = false
	var t *T
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	if t != nil {
		n.Val = *t
		n.NotNil = true
		return nil
	}
	n.Val = *new(T)
	return nil
}

var _ bson.ValueMarshaler = (*Nil[any])(nil)

// MarshalBSONValue implements the bson.ValueMarshaler interface.
func (n Nil[T]) MarshalBSONValue() (byte, []byte, error) {
	if !n.NotNil {
		return byte(bson.TypeNull), []byte{}, nil
	}
	t, data, err := bson.MarshalValue(n.Val)
	return byte(t), data, err
}

var _ bson.ValueUnmarshaler = (*Nil[any])(nil)

// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface.
func (n *Nil[T]) UnmarshalBSONValue(t byte, data []byte) error {
	n.NotNil = false
	if bson.Type(t) == bson.TypeNull {
		n.Val = *new(T)
		return nil
	}
	if err := bson.UnmarshalValue(bson.Type(t), data, &n.Val); err != nil {
		return err
	}
	n.NotNil = true
	return nil
}
