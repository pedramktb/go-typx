package typx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Duration is a struct that represents a duration of time, including both a time.Duration and additional fields for days and months.
// This allows for more flexible time calculations that can account for varying lengths of days and months.
type Duration struct {
	Time  time.Duration // The time.Duration field represents the time part of duration in terms of hours, minutes, seconds, etc.
	Day   int64         // The Day field represents the number of days in the duration. This is useful for calculations that need to account for varying lengths of days (e.g., due to daylight saving time changes).
	Month int64         // The Month field represents the number of months in the duration. This is useful for calculations that need to account for varying lengths of months (e.g., due to the number of days in each month).
}

// String returns a string representation of the Duration, combining the time, day, and month components into a human-readable format.
func (d Duration) String() string {
	return fmt.Sprintf("%dM%dD%s", d.Month, d.Day, d.Time)
}

// Abs returns a new Duration with the absolute values of the time, day, and month components. This is useful for ensuring that all components of the duration are non-negative, regardless of their original signs.
func (d Duration) Abs() Duration {
	return Duration{
		Time:  d.Time.Abs(),
		Day:   max(d.Day, -d.Day),
		Month: max(d.Month, -d.Month),
	}
}

// Add adds another Duration to the current Duration, combining their time, day, and month components.
// The operation is similar to a vector addition, as they are independent unless a point in time is being calculated.
func (d Duration) Add(other Duration) Duration {
	return Duration{
		Time:  d.Time + other.Time,
		Day:   d.Day + other.Day,
		Month: d.Month + other.Month,
	}
}

// Sub subtracts another Duration from the current Duration, combining their time, day, and month components.
// The operation is similar to a vector subtraction, as they are independent unless a point in time is being calculated.
func (d Duration) Sub(other Duration) Duration {
	return Duration{
		Time:  d.Time - other.Time,
		Day:   d.Day - other.Day,
		Month: d.Month - other.Month,
	}
}

var iso8601DurationRe = regexp.MustCompile(
	`^(-?)P(?:(-?\d+)Y)?(?:(-?\d+)M)?(?:(-?\d+)D)?(?:T(?:(-?\d+)H)?(?:(-?\d+)M)?(?:(-?\d+(?:\.\d+)?)S)?)?$`,
)

// MarshalText implements [encoding.TextMarshaler].
// Produces an ISO 8601 duration string (e.g. "P1Y2M3DT4H5M6.789S").
// Month values >= 12 are expressed as years + remaining months.
func (d Duration) MarshalText() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('P')

	years := d.Month / 12
	months := d.Month % 12
	if years != 0 {
		fmt.Fprintf(&b, "%dY", years)
	}
	if months != 0 {
		fmt.Fprintf(&b, "%dM", months)
	}
	if d.Day != 0 {
		fmt.Fprintf(&b, "%dD", d.Day)
	}

	td := d.Time
	hours := td / time.Hour
	td -= hours * time.Hour
	mins := td / time.Minute
	td -= mins * time.Minute

	if hours != 0 || mins != 0 || td != 0 {
		b.WriteByte('T')
		if hours != 0 {
			fmt.Fprintf(&b, "%dH", int64(hours))
		}
		if mins != 0 {
			fmt.Fprintf(&b, "%dM", int64(mins))
		}
		if td != 0 {
			ns := int64(td)
			sign := ""
			if ns < 0 {
				sign = "-"
				ns = -ns
			}
			secs := ns / 1_000_000_000
			frac := ns % 1_000_000_000
			if frac == 0 {
				fmt.Fprintf(&b, "%s%dS", sign, secs)
			} else {
				fracStr := strings.TrimRight(fmt.Sprintf("%09d", frac), "0")
				fmt.Fprintf(&b, "%s%d.%sS", sign, secs, fracStr)
			}
		}
	}

	if b.Len() == 1 { // just "P" — zero duration
		b.WriteString("T0S")
	}
	return []byte(b.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
// Accepts an ISO 8601 duration string (e.g. "P1Y2M3DT4H5M6.789S").
// Years are converted to months (1Y = 12M).
// A leading "-P" negates all components.
func (d *Duration) UnmarshalText(text []byte) error {
	s := string(text)
	m := iso8601DurationRe.FindStringSubmatch(s)
	if m == nil {
		return fmt.Errorf("typx.Duration.UnmarshalText: cannot parse %q as ISO 8601 duration", s)
	}

	sign := int64(1)
	if m[1] == "-" {
		sign = -1
	}

	parseInt64 := func(s string) int64 {
		if s == "" {
			return 0
		}
		n, _ := strconv.ParseInt(s, 10, 64)
		return n
	}

	years := parseInt64(m[2])
	months := parseInt64(m[3])
	days := parseInt64(m[4])
	hours := parseInt64(m[5])
	mins := parseInt64(m[6])

	var secsDur time.Duration
	if m[7] != "" {
		f, err := strconv.ParseFloat(m[7], 64)
		if err != nil {
			return fmt.Errorf("typx.Duration.UnmarshalText: cannot parse seconds %q: %w", m[7], err)
		}
		secsDur = time.Duration(f * float64(time.Second))
	}

	d.Month = sign * (years*12 + months)
	d.Day = sign * days
	d.Time = time.Duration(sign) * (time.Duration(hours)*time.Hour +
		time.Duration(mins)*time.Minute +
		secsDur)
	return nil
}
