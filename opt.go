package typx

// Opt is a type that can be used to represent an optional value.
// It should be used for fields that are optional and might not be present.
// For nullable values, use Nil[T] instead.
type Opt[T any] struct {
	Val T    `json:"val" bson:"val"`
	Set bool `json:"set" bson:"set"`
}

// OptFrom creates an Opt[T] from a non-nil value.
func OptFrom[T any](value T) Opt[T] {
	return Opt[T]{Val: value, Set: true}
}

// OptFromPtr creates an Opt[T] from a pointer. If the pointer is nil, Set is false.
func OptFromPtr[T any](ptr *T) Opt[T] {
	if ptr == nil {
		return Opt[T]{Set: false}
	}
	return Opt[T]{Val: *ptr, Set: true}
}
