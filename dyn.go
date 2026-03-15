package typx

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Dyn is a dynamic type that can hold any value type.
// When using with SQL, the column should be a type that can hold JSON data (JSONB, JSON, TEXT, etc).
type Dyn struct{ Val any }

// MarshalJSON implements the json.Marshaler interface.
func (d Dyn) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Val)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Dyn) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &d.Val)
}

// Scan implements the sql.Scanner interface.
var _ sql.Scanner = (*Dyn)(nil)

func (d *Dyn) Scan(src any) error {
	if src == nil {
		d.Val = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, &d.Val)
	case string:
		return json.Unmarshal([]byte(v), &d.Val)
	}
	return fmt.Errorf("typx.Dyn.Scan: %T is not a string or []byte", src)
}

// Value implements the driver.Valuer interface.
func (d Dyn) Value() (driver.Value, error) {
	return json.Marshal(d.Val)
}

// MarshalBSONValue implements the bson.ValueMarshaler interface.
func (d Dyn) MarshalBSONValue() (byte, []byte, error) {
	t, data, err := bson.MarshalValue(d.Val)
	return byte(t), data, err
}

// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface.
func (d *Dyn) UnmarshalBSONValue(t byte, data []byte) error {
	if err := bson.UnmarshalValue(bson.Type(t), data, &d.Val); err != nil {
		return err
	}
	d.Val = convertBSONToNative(d.Val)
	return nil
}

// convertBSONToNative converts BSON primitive types to Go native types.
func convertBSONToNative(v any) any {
	switch val := v.(type) {
	case bson.D:
		m := make(map[string]any, len(val))
		for _, e := range val {
			m[e.Key] = convertBSONToNative(e.Value)
		}
		return m
	case bson.A:
		arr := make([]any, len(val))
		for i, item := range val {
			arr[i] = convertBSONToNative(item)
		}
		return arr
	case bson.M:
		m := make(map[string]any, len(val))
		for k, item := range val {
			m[k] = convertBSONToNative(item)
		}
		return m
	default:
		return v
	}
}

// The following implementations are provided for convenience,
// but they require that the underlying type implements the respective interfaces.

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (d Dyn) MarshalBinary() ([]byte, error) {
	if marshaler, ok := d.Val.(encoding.BinaryMarshaler); ok {
		return marshaler.MarshalBinary()
	}
	return nil, fmt.Errorf("typx.Dyn.MarshalBinary: %T does not implement encoding.BinaryMarshaler", d.Val)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (d *Dyn) UnmarshalBinary(data []byte) error {
	if unmarshaler, ok := d.Val.(encoding.BinaryUnmarshaler); ok {
		return unmarshaler.UnmarshalBinary(data)
	}
	return fmt.Errorf("typx.Dyn.UnmarshalBinary: %T does not implement encoding.BinaryUnmarshaler", d.Val)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (d Dyn) MarshalText() ([]byte, error) {
	if marshaler, ok := d.Val.(encoding.TextMarshaler); ok {
		return marshaler.MarshalText()
	}
	return nil, fmt.Errorf("typx.Dyn.MarshalText: %T does not implement encoding.TextMarshaler", d.Val)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Dyn) UnmarshalText(data []byte) error {
	if unmarshaler, ok := d.Val.(encoding.TextUnmarshaler); ok {
		return unmarshaler.UnmarshalText(data)
	}
	return fmt.Errorf("typx.Dyn.UnmarshalText: %T does not implement encoding.TextUnmarshaler", d.Val)
}
