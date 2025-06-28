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
type Nil[T any] struct {
	Val    T
	NotNil bool
}

func NilFrom[T any](value T) Nil[T] { return Nil[T]{Val: value, NotNil: true} }

func NilFromPtr[T any](value *T) Nil[T] {
	if value == nil {
		return Nil[T]{}
	}
	return Nil[T]{Val: *value, NotNil: true}
}

func (n Nil[T]) Ptr() *T {
	if !n.NotNil {
		return nil
	}
	return &n.Val
}

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
	var ok bool
	if n.Val, ok = src.(T); ok {
		n.NotNil = true
		return nil
	}
	if b, ok := src.([]byte); ok {
		if n.Val, ok = any(string(b)).(T); ok {
			n.NotNil = true
			return nil
		}
	}
	if s, ok := src.(string); ok {
		if n.Val, ok = any([]byte(s)).(T); ok {
			n.NotNil = true
			return nil
		}
	}
	return fmt.Errorf("cannot scan %v into Nil[%T]", src, n.Val)
}

func (n Nil[T]) Value() (driver.Value, error) {
	if !n.NotNil {
		return nil, nil
	}
	if valuer, ok := any(n.Val).(driver.Valuer); ok {
		return valuer.Value()
	}
	return n.Val, nil
}

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
	return nil, fmt.Errorf("cannot marshal %T into binary", n.Val)
}

func (n *Nil[T]) UnmarshalBinary(data []byte) error {
	n.NotNil = false
	if data == nil {
		n.Val = *new(T)
		return nil
	}
	if unmarshaler, ok := any(n.Val).(encoding.BinaryUnmarshaler); ok {
		err := unmarshaler.UnmarshalBinary(data)
		if err != nil {
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
	return fmt.Errorf("cannot unmarshal binary into %T", n.Val)
}

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
	return nil, fmt.Errorf("cannot marshal %T as text", n.Val)
}

func (n *Nil[T]) UnmarshalText(data []byte) error {
	n.NotNil = false
	if data == nil {
		n.Val = *new(T)
		return nil
	}
	if unmarshaler, ok := any(n.Val).(encoding.TextUnmarshaler); ok {
		err := unmarshaler.UnmarshalText(data)
		if err != nil {
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
	return fmt.Errorf("cannot unmarshal text as %T", n.Val)
}

func (n Nil[T]) MarshalJSON() ([]byte, error) {
	if !n.NotNil {
		return []byte("null"), nil
	}
	return json.Marshal(n.Val)
}

func (n *Nil[T]) UnmarshalJSON(data []byte) error {
	n.NotNil = false
	var t *T
	err := json.Unmarshal(data, &t)
	if err != nil {
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

func (n Nil[T]) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if !n.NotNil {
		var t *T
		return bson.MarshalValue(t)
	}
	return bson.MarshalValue(n.Val)
}

func (n *Nil[T]) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	n.NotNil = false
	if t == bson.TypeNull {
		n.Val = *new(T)
		return nil
	}
	err := bson.UnmarshalValue(t, data, &n.Val)
	if err != nil {
		return err
	}
	n.NotNil = true
	return nil
}
