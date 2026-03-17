package typx

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

/*
DateTime is a [time.Time] wrapper for use as an [OrderedRange] / [OrderedMultiRange] with tsrange, tstzrange, tsmultirange, and tstzmultirange backing types.
All constructors and methods strip the monotonic clock reading so
that two DateTime values representing the same instant compare and marshal identically.

This type also supports operations with the Duration type, which supports dynamic Month and Day components.
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
// A nil value (SQL NULL) zeroes the receiver.
// Use Nil[DateTime] for a type that can be nil instead of zero.
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

// Add adds a Duration with dynamic Month and Day components to the DateTime, returning a new DateTime.
// For the standard time.Duration addition, use the time.Time.Add method on the embedded Time field directly.
func (t DateTime) Add(d Duration) DateTime {
	// Add the time component using time.Time's Add method
	newTime := t.Time.Add(d.Time)

	// Add the day and month components using time.Time's AddDate method
	newTime = newTime.AddDate(0, int(d.Month), int(d.Day))

	return DateTime{newTime}
}

// Sub returns the Duration d such that other.Add(d) == t.
// For getting the standard time.Duration difference between two DateTime values, use the time.Time.Sub(other.Time) method on the embedded Time fields directly.
func (t DateTime) Sub(other DateTime) Duration {
	month := int64(12*(t.Year()-other.Year()) + int(t.Month()) - int(other.Month()))
	day := int64(t.Day() - other.Day())
	// Compute the intermediate time after applying the calendar components, then
	// capture the remaining nanosecond difference as the Time component.
	intermediate := other.Time.AddDate(0, int(month), int(day))
	return Duration{
		Time:  t.Time.Sub(intermediate),
		Day:   day,
		Month: month,
	}
}
