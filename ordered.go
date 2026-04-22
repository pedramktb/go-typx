package typx

// Ordered is a type constraint that matches any type that implements a total ordering via the Compare method.
// Types implementing Ordered can use the OrderedRange and OrderedMultiRange types in this package.
// Additionally, if an Ordered type implements PGBoundMarshaler and PGBoundUnmarshaler,
// it can be used with PostgreSQL range types in a way that produces native range literals instead of JSON-encoded values.
type Ordered[T any] interface {
	Compare(T) int
}

// PGBoundMarshaler is implemented by values that can serialize themselves as a PostgreSQL
// range bound literal (e.g. "2024-01-01 00:00:00+00" for tstzrange, "12345" for numrange).
// OrderedRange.Value and OrderedMultiRange.Value use this to produce a native PG range
// literal; types that don't implement it are JSON-encoded instead.
// In non-PostgreSQL but still SQL contexts, the same PostgreSQL range literal syntax is used but without relying on the database to parse it.
type PGBoundMarshaler interface {
	MarshalPGBound() ([]byte, error)
}

// PGBoundUnmarshaler is implemented by values that can deserialize themselves from a
// PostgreSQL range bound literal.
// OrderedRange.Scan and OrderedMultiRange.Scan use this to parse a native PG range
// literal; types that don't implement it are JSON-decoded instead.
// In non-PostgreSQL but still SQL contexts, the same PostgreSQL range literal syntax is used but without relying on the database to parse it.
type PGBoundUnmarshaler interface {
	UnmarshalPGBound([]byte) error
}
