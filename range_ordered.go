package typx

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// OrderedBound represents one end of an OrderedRange with its value and whether it is exclusive.
// The zero value is inclusive (Exclusive == false) and bounded.
// When Unbounded is true the bound is ±∞; Val and Exclusive are ignored.
type OrderedBound[O Ordered[O]] struct {
	Val       O    `json:"val" bson:"val"`                                 // the bound value
	Exclusive bool `json:"exclusive,omitempty" bson:"exclusive,omitempty"` // whether the bound is exclusive (true) or inclusive (false); ignored when Unbounded is true
	Unbounded bool `json:"unbounded,omitempty" bson:"unbounded,omitempty"` // whether this is an infinite/unbounded bound (true) or a finite bound (false)
}

// OrderedRange supports any Ordered type for in-process use, JSON, and BSON storage.
// The Scan and Value methods implement PostgreSQL range literals; note that string-based
// types have no corresponding PostgreSQL range type and should not be used with SQL.
// In the case of a type implementing PGBoundUnmarshaler/Marshaler, those are used to parse/format the bounds of the range literal.
type OrderedRange[O Ordered[O]] struct {
	Lower OrderedBound[O] `json:"lower" bson:"lower"` // lower bound
	Upper OrderedBound[O] `json:"upper" bson:"upper"` // upper bound
}

func (r OrderedRange[O]) Contains(val O) bool {
	var lowerOk bool
	if r.Lower.Unbounded {
		lowerOk = true
	} else {
		lo := r.Lower.Val.Compare(val)
		lowerOk = lo < 0 || (lo == 0 && !r.Lower.Exclusive)
	}
	var upperOk bool
	if r.Upper.Unbounded {
		upperOk = true
	} else {
		hi := r.Upper.Val.Compare(val)
		upperOk = hi > 0 || (hi == 0 && !r.Upper.Exclusive)
	}
	return lowerOk && upperOk
}

// Scan implements the sql.Scanner interface for OrderedRange.
// If *O implements PGBoundUnmarshaler it parses a PostgreSQL range literal (e.g. [min,max]).
// Otherwise it JSON-decodes a range object, suitable for text/jsonb columns.
func (r *OrderedRange[O]) Scan(src any) error {
	var str string
	switch v := src.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("typx.OrderedRange.Scan: %T is not a string or []byte", src)
	}
	var zero O
	if _, ok := any(&zero).(PGBoundUnmarshaler); ok {
		return r.scanRangeLiteral(str)
	}
	return json.Unmarshal([]byte(str), r)
}

func (r *OrderedRange[O]) scanRangeLiteral(str string) error {
	if str == "empty" {
		*r = OrderedRange[O]{}
		return nil
	}
	if len(str) < 3 {
		return fmt.Errorf("typx.OrderedRange.Scan: %q is too short to be a valid range literal", str)
	}
	lowerExclusive := str[0] == '('
	upperExclusive := str[len(str)-1] == ')'
	inner := str[1 : len(str)-1]
	before, after, ok := strings.Cut(inner, ",")
	if !ok {
		return fmt.Errorf("typx.OrderedRange.Scan: %q is not a valid range literal", str)
	}
	if before == "" {
		// empty lower bound means unbounded (-∞); bracket convention is always '('
		r.Lower = OrderedBound[O]{Unbounded: true}
	} else {
		r.Lower.Unbounded = false
		r.Lower.Exclusive = lowerExclusive
		if err := any(&r.Lower.Val).(PGBoundUnmarshaler).UnmarshalPGBound([]byte(before)); err != nil {
			return fmt.Errorf("typx.OrderedRange.Scan: parsing lower bound %q: %w", before, err)
		}
	}
	if after == "" {
		// empty upper bound means unbounded (+∞); bracket convention is always ')'
		r.Upper = OrderedBound[O]{Unbounded: true}
	} else {
		r.Upper.Unbounded = false
		r.Upper.Exclusive = upperExclusive
		if err := any(&r.Upper.Val).(PGBoundUnmarshaler).UnmarshalPGBound([]byte(after)); err != nil {
			return fmt.Errorf("typx.OrderedRange.Scan: parsing upper bound %q: %w", after, err)
		}
	}
	return nil
}

// Value implements the driver.Valuer interface for OrderedRange.
// If O or *O implements PGBoundMarshaler it returns a PostgreSQL range literal (e.g. [min,max)).
// Otherwise it JSON-encodes the range for storage in a text/jsonb column.
func (r OrderedRange[O]) Value() (driver.Value, error) {
	var zero O
	if _, ok := any(zero).(PGBoundMarshaler); ok {
		return r.valueRangeLiteral()
	}
	if _, ok := any(&zero).(PGBoundMarshaler); ok {
		return r.valueRangeLiteralPtr()
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("typx.OrderedRange.Value: %w", err)
	}
	return string(b), nil
}

func (r OrderedRange[O]) valueRangeLiteral() (driver.Value, error) {
	var lo, hi, lowerStr, upperStr string
	if r.Lower.Unbounded {
		lo = "("
		lowerStr = ""
	} else {
		if r.Lower.Exclusive {
			lo = "("
		} else {
			lo = "["
		}
		b, err := any(r.Lower.Val).(PGBoundMarshaler).MarshalPGBound()
		if err != nil {
			return nil, fmt.Errorf("typx.OrderedRange.Value: marshaling lower bound: %w", err)
		}
		lowerStr = string(b)
	}
	if r.Upper.Unbounded {
		hi = ")"
		upperStr = ""
	} else {
		if r.Upper.Exclusive {
			hi = ")"
		} else {
			hi = "]"
		}
		b, err := any(r.Upper.Val).(PGBoundMarshaler).MarshalPGBound()
		if err != nil {
			return nil, fmt.Errorf("typx.OrderedRange.Value: marshaling upper bound: %w", err)
		}
		upperStr = string(b)
	}
	return fmt.Sprintf("%s%s,%s%s", lo, lowerStr, upperStr, hi), nil
}

// valueRangeLiteralPtr is like valueRangeLiteral but calls MarshalPGBound via a pointer receiver.
func (r OrderedRange[O]) valueRangeLiteralPtr() (driver.Value, error) {
	var lo, hi, lowerStr, upperStr string
	if r.Lower.Unbounded {
		lo = "("
		lowerStr = ""
	} else {
		if r.Lower.Exclusive {
			lo = "("
		} else {
			lo = "["
		}
		b, err := any(&r.Lower.Val).(PGBoundMarshaler).MarshalPGBound()
		if err != nil {
			return nil, fmt.Errorf("typx.OrderedRange.Value: marshaling lower bound: %w", err)
		}
		lowerStr = string(b)
	}
	if r.Upper.Unbounded {
		hi = ")"
		upperStr = ""
	} else {
		if r.Upper.Exclusive {
			hi = ")"
		} else {
			hi = "]"
		}
		b, err := any(&r.Upper.Val).(PGBoundMarshaler).MarshalPGBound()
		if err != nil {
			return nil, fmt.Errorf("typx.OrderedRange.Value: marshaling upper bound: %w", err)
		}
		upperStr = string(b)
	}
	return fmt.Sprintf("%s%s,%s%s", lo, lowerStr, upperStr, hi), nil
}

type OrderedMultiRange[O Ordered[O]] []OrderedRange[O]

func orderedRangeOverlaps[O Ordered[O]](current, r OrderedRange[O]) bool {
	// current extends to +∞: every subsequent range overlaps.
	if current.Upper.Unbounded {
		return true
	}
	// r starts at -∞: it always overlaps any current range.
	if r.Lower.Unbounded {
		return true
	}
	c := r.Lower.Val.Compare(current.Upper.Val)
	if c < 0 {
		return true
	}
	if c == 0 {
		return !r.Lower.Exclusive || !current.Upper.Exclusive
	}
	return false
}

func mergeOrderedUpper[O Ordered[O]](a, b OrderedBound[O]) OrderedBound[O] {
	if b.Unbounded {
		return b // +∞ is the furthest possible upper bound
	}
	if a.Unbounded {
		return a
	}
	c := b.Val.Compare(a.Val)
	if c > 0 {
		return b
	}
	if c == 0 && !b.Exclusive {
		return b
	}
	return a
}

func NewOrderedMultiRange[O Ordered[O]](src []OrderedRange[O]) OrderedMultiRange[O] {
	if len(src) == 0 {
		return nil
	}
	rs := make(OrderedMultiRange[O], len(src))
	copy(rs, src)
	// Sort by lower bound value, preferring inclusive lower bounds first.
	// Unbounded lower (-∞) always sorts before any finite lower bound.
	slices.SortFunc(rs, func(a, b OrderedRange[O]) int {
		switch {
		case a.Lower.Unbounded && b.Lower.Unbounded:
			return 0
		case a.Lower.Unbounded:
			return -1
		case b.Lower.Unbounded:
			return 1
		}
		if c := a.Lower.Val.Compare(b.Lower.Val); c != 0 {
			return c
		}
		if !a.Lower.Exclusive && b.Lower.Exclusive {
			return -1
		}
		if a.Lower.Exclusive && !b.Lower.Exclusive {
			return 1
		}
		return 0
	})
	// Merge overlapping or adjacent ranges.
	mr := make(OrderedMultiRange[O], 0, len(rs))
	current := rs[0]
	for _, r := range rs[1:] {
		if orderedRangeOverlaps(current, r) {
			current.Upper = mergeOrderedUpper(current.Upper, r.Upper)
		} else {
			mr = append(mr, current)
			current = r
		}
	}
	mr = append(mr, current)
	return mr
}

func (mr OrderedMultiRange[O]) Contains(val O) bool {
	for _, r := range mr {
		if r.Contains(val) {
			return true
		}
	}
	return false
}

// Scan implements the sql.Scanner interface for OrderedMultiRange.
// If *O implements PGBoundUnmarshaler it parses a PostgreSQL multirange literal (e.g. {[1,5],[10,20]}).
// Otherwise it JSON-decodes a JSON array of range objects, suitable for text/jsonb columns.
func (mr *OrderedMultiRange[O]) Scan(src any) error {
	var str string
	switch v := src.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("typx.OrderedMultiRange.Scan: %T is not a string or []byte", src)
	}
	var zero O
	if _, ok := any(&zero).(PGBoundUnmarshaler); ok {
		return mr.scanMultiRangeLiteral(str)
	}
	return json.Unmarshal([]byte(str), mr)
}

func (mr *OrderedMultiRange[O]) scanMultiRangeLiteral(str string) error {
	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
		return fmt.Errorf("typx.OrderedMultiRange.Scan: %q is too short to be a valid multirange literal", str)
	}
	inner := str[1 : len(str)-1]
	if inner == "" {
		*mr = OrderedMultiRange[O]{}
		return nil
	}
	// Split at the boundary between consecutive ranges: the ',' that follows a closing bracket.
	var parts []string
	start := 0
	for i := 1; i < len(inner); i++ {
		if (inner[i-1] == ']' || inner[i-1] == ')') && inner[i] == ',' {
			parts = append(parts, inner[start:i])
			start = i + 1
		}
	}
	parts = append(parts, inner[start:])
	result := make(OrderedMultiRange[O], 0, len(parts))
	for _, part := range parts {
		var r OrderedRange[O]
		if err := r.scanRangeLiteral(part); err != nil {
			return fmt.Errorf("typx.OrderedMultiRange.Scan: parsing range %q: %w", part, err)
		}
		result = append(result, r)
	}
	*mr = result
	return nil
}

// Value implements the driver.Valuer interface for OrderedMultiRange.
// If O or *O implements PGBoundMarshaler it returns a PostgreSQL multirange literal (e.g. {[1,5],[10,20]}).
// Otherwise it JSON-encodes the multirange as a JSON array for storage in a text/jsonb column.
func (mr OrderedMultiRange[O]) Value() (driver.Value, error) {
	var zero O
	if _, ok := any(zero).(PGBoundMarshaler); ok {
		return mr.valueMultiRangeLiteral()
	}
	if _, ok := any(&zero).(PGBoundMarshaler); ok {
		return mr.valueMultiRangeLiteralPtr()
	}
	b, err := json.Marshal([]OrderedRange[O](mr))
	if err != nil {
		return nil, fmt.Errorf("typx.OrderedMultiRange.Value: %w", err)
	}
	return string(b), nil
}

func (mr OrderedMultiRange[O]) valueMultiRangeLiteral() (driver.Value, error) {
	parts := make([]string, 0, len(mr))
	for _, r := range mr {
		v, err := r.valueRangeLiteral()
		if err != nil {
			return nil, err
		}
		parts = append(parts, v.(string))
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

func (mr OrderedMultiRange[O]) valueMultiRangeLiteralPtr() (driver.Value, error) {
	parts := make([]string, 0, len(mr))
	for _, r := range mr {
		v, err := r.valueRangeLiteralPtr()
		if err != nil {
			return nil, err
		}
		parts = append(parts, v.(string))
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}
