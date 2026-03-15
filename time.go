package typx

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

/*
DateTime is a [time.Time] wrapper for use as an [OrderedRange] / [OrderedMultiRange]
All constructors and methods strip the monotonic clock reading so
that two DateTime values representing the same instant compare and marshal identically.
*/
type DateTime struct{ time.Time }

func (t DateTime) Compare(other DateTime) int {
	return t.Time.Compare(other.Time)
}

// MarshalPGBound implements [PGBoundMarshaler].
// Produces a PostgreSQL timestamptz literal.
func (t DateTime) MarshalPGBound() ([]byte, error) {
	return []byte(t.Time.Round(0).Format(time.RFC3339Nano)), nil
}

// UnmarshalPGBound implements [PGBoundUnmarshaler].
// Accepts a PostgreSQL timestamptz literal, including the double-quoted form
// that PostgreSQL uses for timestamp bounds inside range literals
func (t *DateTime) UnmarshalPGBound(text []byte) error {
	s := strings.Trim(string(text), `"`)
	for _, layout := range []string{
		// PostgreSQL native output (space separator, Z or ±HH:MM offset)
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05Z07:00",
		// PostgreSQL native output with bare ±HH offset (e.g. +00 for UTC)
		"2006-01-02 15:04:05.999999999-07",
		"2006-01-02 15:04:05-07",
		// timestamp without time zone (no offset)
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed.UTC()
			return nil
		}
	}
	return fmt.Errorf("typx.DateTime.UnmarshalPGBound: cannot parse %q", s)
}

// Value implements [driver.Valuer].
// Returns the time with the monotonic clock stripped.
func (t DateTime) Value() (driver.Value, error) {
	return t.Time.Round(0), nil
}

// Scan implements [sql.Scanner].
// The driver is expected to deliver a [time.Time] value (which pgx and database/sql
// do for DATE, TIMESTAMP, and TIMESTAMPTZ columns alike).
func (t *DateTime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		t.Time = v.Round(0)
	case *time.Time:
		if v != nil {
			t.Time = v.Round(0)
		}
	default:
		return fmt.Errorf("typx.DateTime.Scan: unsupported type %T", value)
	}
	return nil
}
