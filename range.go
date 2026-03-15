package typx

import (
	"cmp"
	"database/sql/driver"
	"fmt"
	"slices"
	"strings"
)

// Bound represents one end of a range with its value and whether it is exclusive.
// The zero value is inclusive (Exclusive == false) and bounded.
// When Unbounded is true the bound is ±∞; Val and Exclusive are ignored.
type Bound[O cmp.Ordered] struct {
	Val       O    `json:"val" bson:"val"`                                 // the bound value
	Exclusive bool `json:"exclusive,omitempty" bson:"exclusive,omitempty"` // whether the bound is exclusive (true) or inclusive (false); ignored when Unbounded is true
	Unbounded bool `json:"unbounded,omitempty" bson:"unbounded,omitempty"` // whether this is an infinite/unbounded bound (true) or a finite bound (false)
}

// Range supports any cmp.Ordered type for in-process use, JSON, and BSON storage.
// The Scan and Value methods implement PostgreSQL range literals; note that string-based
// types have no corresponding PostgreSQL range type and should not be used with SQL.
type Range[O cmp.Ordered] struct {
	Lower Bound[O] `json:"lower" bson:"lower"` // lower bound
	Upper Bound[O] `json:"upper" bson:"upper"` // upper bound
}

func (r Range[O]) Contains(val O) bool {
	var lowerOk bool
	if r.Lower.Unbounded {
		lowerOk = true
	} else {
		lo := cmp.Compare(val, r.Lower.Val)
		lowerOk = lo > 0 || (lo == 0 && !r.Lower.Exclusive)
	}
	var upperOk bool
	if r.Upper.Unbounded {
		upperOk = true
	} else {
		hi := cmp.Compare(val, r.Upper.Val)
		upperOk = hi < 0 || (hi == 0 && !r.Upper.Exclusive)
	}
	return lowerOk && upperOk
}

// Scan implements the sql.Scanner interface for Range.
// It expects a PostgreSQL range literal such as [min,max] or (min,max).
func (r *Range[O]) Scan(src any) error {
	var str string
	switch v := src.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("typx.Range.Scan: %T is not a string or []byte", src)
	}
	if str == "empty" {
		*r = Range[O]{}
		return nil
	}
	// Expect at least bracket + comma + bracket: (,)
	if len(str) < 3 {
		return fmt.Errorf("typx.Range.Scan: %q is too short to be a valid range literal", str)
	}
	lowerExclusive := str[0] == '('
	upperExclusive := str[len(str)-1] == ')'
	inner := str[1 : len(str)-1]
	before, after, ok := strings.Cut(inner, ",")
	if !ok {
		return fmt.Errorf("typx.Range.Scan: %q is not a valid range literal", str)
	}
	if before == "" {
		// empty lower bound means unbounded (-∞); bracket convention is always '('
		r.Lower = Bound[O]{Unbounded: true}
	} else {
		r.Lower.Unbounded = false
		r.Lower.Exclusive = lowerExclusive
		if _, err := fmt.Sscan(before, &r.Lower.Val); err != nil {
			return fmt.Errorf("typx.Range.Scan: parsing lower bound %q: %w", before, err)
		}
	}
	if after == "" {
		// empty upper bound means unbounded (+∞); bracket convention is always ')'
		r.Upper = Bound[O]{Unbounded: true}
	} else {
		r.Upper.Unbounded = false
		r.Upper.Exclusive = upperExclusive
		if _, err := fmt.Sscan(after, &r.Upper.Val); err != nil {
			return fmt.Errorf("typx.Range.Scan: parsing upper bound %q: %w", after, err)
		}
	}
	return nil
}

// Value implements the driver.Valuer interface for Range.
// It returns a PostgreSQL range literal respecting the bound types, e.g. [min,max) or (min,max].
func (r Range[O]) Value() (driver.Value, error) {
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
		lowerStr = fmt.Sprintf("%v", r.Lower.Val)
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
		upperStr = fmt.Sprintf("%v", r.Upper.Val)
	}
	return fmt.Sprintf("%s%s,%s%s", lo, lowerStr, upperStr, hi), nil
}

type MultiRange[O cmp.Ordered] []Range[O]

// rangeOverlaps reports whether r (whose lower bound is >= current's lower bound after sorting)
// overlaps or is adjacent to current.
func rangeOverlaps[O cmp.Ordered](current, r Range[O]) bool {
	// current extends to +∞: every subsequent range overlaps.
	if current.Upper.Unbounded {
		return true
	}
	// r starts at -∞: it always overlaps any current range.
	if r.Lower.Unbounded {
		return true
	}
	c := cmp.Compare(r.Lower.Val, current.Upper.Val)
	if c < 0 {
		return true
	}
	// Adjacent at the same value: overlaps only when at least one side is inclusive.
	if c == 0 {
		return !r.Lower.Exclusive || !current.Upper.Exclusive
	}
	return false
}

// mergeUpper returns the further of two upper bounds, preferring inclusive when equal.
func mergeUpper[O cmp.Ordered](a, b Bound[O]) Bound[O] {
	if b.Unbounded {
		return b // +∞ is the furthest possible upper bound
	}
	if a.Unbounded {
		return a
	}
	c := cmp.Compare(b.Val, a.Val)
	if c > 0 {
		return b
	}
	if c == 0 && !b.Exclusive {
		return b // inclusive beats exclusive
	}
	return a
}

// NewMultiRange takes a slice of ranges and returns a new multi range that is optimized by merging overlapping ranges.
func NewMultiRange[O cmp.Ordered](src []Range[O]) MultiRange[O] {
	if len(src) == 0 {
		return nil
	}
	rs := make(MultiRange[O], len(src))
	copy(rs, src)
	// Sort by lower bound value, then prefer inclusive lower bounds first.
	// Unbounded lower (-∞) always sorts before any finite lower bound.
	slices.SortFunc(rs, func(a, b Range[O]) int {
		switch {
		case a.Lower.Unbounded && b.Lower.Unbounded:
			return 0
		case a.Lower.Unbounded:
			return -1
		case b.Lower.Unbounded:
			return 1
		}
		if c := cmp.Compare(a.Lower.Val, b.Lower.Val); c != 0 {
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
	mr := make(MultiRange[O], 0, len(rs))
	current := rs[0]
	for _, r := range rs[1:] {
		if rangeOverlaps(current, r) {
			current.Upper = mergeUpper(current.Upper, r.Upper)
		} else {
			mr = append(mr, current)
			current = r
		}
	}
	mr = append(mr, current)
	return mr
}

func (mr MultiRange[O]) Contains(val O) bool {
	for _, r := range mr {
		if r.Contains(val) {
			return true
		}
	}
	return false
}

// Scan implements the sql.Scanner interface for MultiRange.
// It expects a PostgreSQL multirange literal such as {[1,5],[10,20]}.
func (mr *MultiRange[O]) Scan(src any) error {
	var str string
	switch v := src.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("typx.MultiRange.Scan: %T is not a string or []byte", src)
	}
	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
		return fmt.Errorf("typx.MultiRange.Scan: %q is too short to be a valid multirange literal", str)
	}
	inner := str[1 : len(str)-1]
	if inner == "" {
		*mr = MultiRange[O]{}
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
	result := make(MultiRange[O], 0, len(parts))
	for _, part := range parts {
		var r Range[O]
		if err := r.Scan(part); err != nil {
			return fmt.Errorf("typx.MultiRange.Scan: parsing range %q: %w", part, err)
		}
		result = append(result, r)
	}
	*mr = result
	return nil
}

// Value implements the driver.Valuer interface for MultiRange.
// It returns a PostgreSQL multirange literal such as {[1,5],[10,20]}.
func (mr MultiRange[O]) Value() (driver.Value, error) {
	parts := make([]string, 0, len(mr))
	for _, r := range mr {
		v, err := r.Value()
		if err != nil {
			return nil, err
		}
		parts = append(parts, v.(string))
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}
