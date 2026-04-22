# <img src="./typx.svg" alt="typX" />

typX is a Go package that provides useful types, helpers and utilities. It currently includes the following:

- **Nil** - Represents a nullable value. Implements SQL, JSON, and BSON encoding. Use instead of `*T` for nullable fields.
- **Opt** - Represents an optional (present vs. absent) value. Implements JSON (`omitzero`) and BSON (`omitempty`) encoding. Use for optional fields in API payloads.
- **Dyn** - A dynamic type that holds any value with full support for JSON, SQL, and BSON encoding. Useful for storing arbitrary JSON/BSON data in databases.
- **KV / KVs** - A generic key-value pair and slice type with full JSON, SQL, and BSON encoding support. Useful for query parameters, metadata, and ordered maps.
- **Range / MultiRange** - PostgreSQL-compatible range and multirange types for any `cmp.Ordered` type (e.g. `int`, `float64`, `string`). Support SQL scan/value using native PG range literals.
- **OrderedRange / OrderedMultiRange** - Range and multirange types for custom types implementing the `Ordered[T]` interface. When the bound type also implements `PGBoundMarshaler` / `PGBoundUnmarshaler`, native PG range literals are used; otherwise values are JSON-encoded.
- **Ordered / PGBoundMarshaler / PGBoundUnmarshaler** - Interfaces for custom ordered types and PostgreSQL bound serialization.
- **DateTime** - A `time.Time` wrapper implementing `Ordered[DateTime]`, `PGBoundMarshaler`, `PGBoundUnmarshaler`, `sql.Scanner`, and `driver.Valuer`. Ready to use as bounds in `OrderedRange[DateTime]` / `OrderedMultiRange[DateTime]` (maps to `tstzrange` / `tstzmultirange`).
- **Str64** - A `[]byte` type that transparently encodes as unpadded base64url (alphabet: `A-Z a-z 0-9 - _`) in JSON and any text transport via `encoding.TextMarshaler` / `encoding.TextUnmarshaler`. Useful for compact UUID representation, URL-safe tokens, and loggable binary values.
- **Must / Must2 / Must3** - Generic panic-on-error helpers for unwrapping one, two, or three return values alongside an `error`.
- **Pointer Helpers** - `FromPtr` and `FromPtrOrZero` for safe pointer dereferencing. (`Ptr` is deprecated since Go 1.26 provides value initialization via the built-in `new`.)
- **Duration** - A duration type that combines a `time.Duration` with separate `Day` and `Month` fields, enabling calendar-aware durations (e.g. PostgreSQL `interval`). Marshals to/from ISO 8601 (`P1Y2M3DT4H5M6S`) via `encoding.TextMarshaler` / `encoding.TextUnmarshaler`.
- **URL** - A `url.URL` wrapper that serialises as a plain URL string in all common transports. Implements `encoding.TextMarshaler` / `encoding.TextUnmarshaler` (JSON), `driver.Valuer`, and `sql.Scanner`. Use `Nil[URL]` for nullable SQL columns.

## Example

```go
// Main model
type User struct {
    ID             typx.Str64                      // compact base64url representation of raw ID bytes
    Name           string
    Mobile         string
    Landline       typx.Nil[string]                // nullable, instead of *string
    AdditionalInfo typx.Dyn                        // arbitrary JSON/BSON data
    ActiveWindow   typx.OrderedRange[typx.DateTime] // tstzrange column
}

// A sample DTO for PATCH /api/v1/user/{id}
// Fields absent from the request body remain unset (Opt.Set == false).
type UpdateUserDTO struct {
    Name           typx.Opt[string]
    Mobile         typx.Opt[string]
    Landline       typx.Opt[typx.Nil[string]]
    AdditionalInfo typx.Opt[typx.Dyn]
}

// KV / KVs example
func NewDatabaseConnection(config ...typx.KV[string, string]) (*sql.DB, error) {
    for _, kv := range config {
        cfg.Set(kv.Key, kv.Val)
    }
    // ...
}

// Range example (built-in ordered types, e.g. int)
var r typx.Range[int]
_ = r.Scan("[10,20)")         // lower inclusive, upper exclusive
contains := r.Contains(15)   // true

// MultiRange example
var mr typx.MultiRange[int]
_ = mr.Scan("{[1,5],[10,20)}")
contains = mr.Contains(12)   // true

// OrderedRange with DateTime (maps to tstzrange)
var tr typx.OrderedRange[typx.DateTime]
_ = tr.Scan(`["2024-01-01 00:00:00+00","2025-01-01 00:00:00+00")`)

// Str64 — compact URL-safe UUID
uid := uuid.New()
s := typx.Str64(uid[:])      // 16 bytes → 22-char base64url string
fmt.Println(s)               // e.g. "BpLnfgDsc2WD8F2qNkHydQ"
s2, _ := typx.FromString(s.String())
uid2, _ := uuid.FromBytes(s2)

// Must helpers
val := typx.Must(strconv.Atoi("42")) // panics on error

// Duration — calendar-aware duration, marshals as ISO 8601
d := typx.Duration{Month: 14, Day: 3, Time: 4*time.Hour + 30*time.Minute}
text, _ := d.MarshalText() // "P1Y2M3DT4H30M"
var d2 typx.Duration
_ = d2.UnmarshalText(text)

// URL — plain-string transport for url.URL
cb, _ := typx.URLFrom("https://example.com/callback?token=abc")
fmt.Println(cb) // https://example.com/callback?token=abc
// Stores as TEXT in SQL; use Nil[URL] for nullable columns
type Webhook struct {
    ID       int
    Callback typx.Nil[typx.URL]
}
```
