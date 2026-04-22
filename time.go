package typx

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

/*
[DateTime] is a [time.Time] wrapper for use as an [OrderedRange] / [OrderedMultiRange] with tsrange, tstzrange, tsmultirange, and tstzmultirange backing types.
All constructors and methods strip the monotonic clock reading so
that two [DateTime] values representing the same instant compare and marshal identically.

This type also supports operations with the [Duration] type, which supports dynamic Month and Day components.
*/
type DateTime struct{ time.Time }

var _ Ordered[DateTime] = DateTime{}

// Compare implements the [Ordered] interface by comparing the time values with the monotonic clock stripped.
func (t DateTime) Compare(other DateTime) int {
	return t.Time.Compare(other.Time)
}

var _ PGBoundMarshaler = DateTime{}

// MarshalPGBound implements [PGBoundMarshaler].
// Produces a PostgreSQL timestamptz literal.
func (t DateTime) MarshalPGBound() ([]byte, error) {
	return []byte(t.Time.Round(0).Format(time.RFC3339Nano)), nil
}

var _ PGBoundUnmarshaler = (*DateTime)(nil)

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

var _ driver.Valuer = DateTime{}

// Value implements [driver.Valuer].
// Returns the time with the monotonic clock stripped.
func (t DateTime) Value() (driver.Value, error) {
	return t.Time.Round(0), nil
}

var _ sql.Scanner = (*DateTime)(nil)

// Scan implements [sql.Scanner].
// The driver is expected to deliver a [time.Time] value (which pgx and database/sql
// do for DATE, TIMESTAMP, and TIMESTAMPTZ columns alike).
// A nil value (SQL NULL) zeroes the receiver.
// Use [Nil][[DateTime]] for a type that can be nil instead of zero.
func (t *DateTime) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		t.Time = time.Time{}
	case time.Time:
		t.Time = v.Round(0)
	case *time.Time:
		if v == nil {
			t.Time = time.Time{}
		} else {
			t.Time = v.Round(0)
		}
	default:
		return fmt.Errorf("typx.DateTime.Scan: unsupported type %T", value)
	}
	return nil
}

// AddDuration adds a [Duration] with dynamic Month and Day components to the [DateTime], returning a new [DateTime].
func (t DateTime) AddDuration(d Duration) DateTime {
	newTime := t.Time.Add(d.Time)
	newTime = newTime.AddDate(0, int(d.Month), int(d.Day))
	return DateTime{newTime.Round(0)}
}

// SubDateTime returns the [Duration] d such that t.Equal(other.AddDuration(d)).
func (t DateTime) SubDateTime(other DateTime) Duration {
	yearDiff := t.Year() - other.Year()
	monthDiff := int(t.Month()) - int(other.Month())
	month := int64(12*yearDiff + monthDiff)
	day := int64(t.Day() - other.Day())
	// Compute the intermediate time after applying the calendar components, then
	// capture the remaining nanosecond difference as the Time component.
	// Use yearDiff/monthDiff separately to avoid a potentially large int64→int cast.
	intermediate := other.Time.AddDate(yearDiff, monthDiff, int(day))
	return Duration{
		Time:  t.Time.Sub(intermediate),
		Day:   day,
		Month: month,
	}
}
