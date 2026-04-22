package typx

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"net/url"
)

/*
[URL] is a [url.URL] wrapper that serialises as a plain URL string in all common transports.

Because it implements [encoding.TextMarshaler] / [encoding.TextUnmarshaler], JSON (and any
other text-based format) represents the value as a quoted URL string with no extra nesting.

It also implements [driver.Valuer] and [sql.Scanner] so it can be stored directly in any
SQL column that holds text (VARCHAR, TEXT, etc.), without manual .String() / url.Parse calls
at the call site:

	type Row struct {
	    ID       int
	    Callback typx.URL
	}

Use [Nil][[URL]] for a nullable SQL column.
*/
type URL struct{ url.URL }

// [URLFrom] parses str and returns a URL or an error.
func URLFrom(str string) (URL, error) {
	u, err := url.Parse(str)
	if err != nil {
		return URL{}, fmt.Errorf("typx.URLFrom: %w", err)
	}
	return URL{*u}, nil
}

var _ fmt.Stringer = URL{}

// String returns the URL in its canonical string form.
func (u URL) String() string {
	return u.URL.String()
}

var _ encoding.TextMarshaler = URL{}

// MarshalText implements [encoding.TextMarshaler].
func (u URL) MarshalText() ([]byte, error) {
	return []byte(u.URL.String()), nil
}

var _ encoding.TextUnmarshaler = (*URL)(nil)

// UnmarshalText implements [encoding.TextUnmarshaler].
func (u *URL) UnmarshalText(text []byte) error {
	parsed, err := url.Parse(string(text))
	if err != nil {
		return fmt.Errorf("typx.URL.UnmarshalText: %w", err)
	}
	u.URL = *parsed
	return nil
}

var _ driver.Valuer = URL{}

// Value implements [driver.Valuer].
// The URL is stored as a plain string in the database.
func (u URL) Value() (driver.Value, error) {
	return u.URL.String(), nil
}

var _ sql.Scanner = (*URL)(nil)

// Scan implements [sql.Scanner].
// Accepts string or []byte from the driver.
// A nil src (SQL NULL) zeroes the receiver.
// Use Nil[[URL]] for a type that can be nil instead of zero.
func (u *URL) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		u.URL = url.URL{}
	case string:
		parsed, err := url.Parse(v)
		if err != nil {
			return fmt.Errorf("typx.URL.Scan: %w", err)
		}
		u.URL = *parsed
	case []byte:
		parsed, err := url.Parse(string(v))
		if err != nil {
			return fmt.Errorf("typx.URL.Scan: %w", err)
		}
		u.URL = *parsed
	default:
		return fmt.Errorf("typx.URL.Scan: unsupported type %T", src)
	}
	return nil
}
