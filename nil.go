package typx

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// Nil is a type that can be used to represent a nil/nullable value.
// It implements most of the interfaces that are used to marshal and unmarshal values.
// For optional values that are not present, use Opt[T] instead.
type Nil[T any] struct {
	Val    T
	NotNil bool
}

// NilFrom creates a Nil[T] from a non-nil value.
func NilFrom[T any](value T) Nil[T] { return Nil[T]{Val: value, NotNil: true} }

// NilFromPtr creates a Nil[T] from a pointer. If the pointer is nil, NotNil is false.
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

// Scan implements the sql.Scanner interface.
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
	return fmt.Errorf("cannot scan %v into Nil[%T]", src, n.Val)
}

// Value implements the driver.Valuer interface.
func (n Nil[T]) Value() (driver.Value, error) {
	if !n.NotNil {
		return nil, nil
	}
	if valuer, ok := any(n.Val).(driver.Valuer); ok {
		return valuer.Value()
	}
	return n.Val, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
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
	return nil, fmt.Errorf("cannot marshal %T into binary: expected encoding.BinaryMarshaler, string, or []byte", n.Val)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
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
	return fmt.Errorf("cannot unmarshal binary into %T: expected encoding.BinaryUnmarshaler, string, or []byte", n.Val)
}

// MarshalText implements the encoding.TextMarshaler interface.
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
	return nil, fmt.Errorf("cannot marshal %T as text: expected encoding.TextMarshaler, string, or []byte", n.Val)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
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
	return fmt.Errorf("cannot unmarshal text as %T: expected encoding.TextUnmarshaler, string, or []byte", n.Val)
}

// MarshalJSON implements the json.Marshaler interface.
func (n Nil[T]) MarshalJSON() ([]byte, error) {
	if !n.NotNil {
		return []byte("null"), nil
	}
	return json.Marshal(n.Val)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
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

// MarshalBSONValue implements the bson.ValueMarshaler interface.
func (n Nil[T]) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if !n.NotNil {
		return bson.MarshalValue(new(T))
	}
	return bson.MarshalValue(n.Val)
}

// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface.
func (n *Nil[T]) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	n.NotNil = false
	if t == bson.TypeNull {
		n.Val = *new(T)
		return nil
	}
	if err := bson.UnmarshalValue(t, data, &n.Val); err != nil {
		return err
	}
	n.NotNil = true
	return nil
}
