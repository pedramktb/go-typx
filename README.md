# typx

typx is a go package that provides useful types, helpers and utilities. It currently includes the following:

- **Nil** - A type that can be used to represent a nil/nullable value. It implements interfaces for SQL, JSON and BSON encoding.
- **Opt** - An optional type that can be used to represent an optional value. Intends to be used for optional fields in JSON payloads instead of null or undefined values.
- **Dyn** - A dynamic type that can hold any value with full support for JSON, SQL, and BSON encoding. Useful for storing arbitrary JSON data in databases.
- **KV(s)** - A simple key-value pair (slice) type that can be used for various purposes, such as query parameters or metadata.
- **Pointer Helpers** - A set of helper functions for working with pointers.

## Example
```go
// Main model
type User struct {
    ID uuid.UUID
    Name string
    Mobile string
    Landline typx.Nil[string] // Instead of *string
    AdditionalInfo typx.Dyn // For storing arbitrary JSON data
}

// A sample DTO for PATCH /api/v1/user/{id}
type UpdateUserDTO struct {
    Name typx.Opt[string]
    Mobile typx.Opt[string]
    Landline typx.Opt[typx.Nil[string]]
    AdditionalInfo typx.Opt[typx.Dyn]
}

// KV example
func NewDatabaseConnection(config ...typx.KV[string, string]) (*sql.DB, error) {
    // ...
    for _, kv := range config {
        cfg.Set(kv.Key, kv.Value)
    }
    // ...
}
```
