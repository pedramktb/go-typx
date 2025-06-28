package typx

// Opt is a type that can be used to represent an optional value
// instead of null or undefined values.
type Opt[T any] struct {
	Val T    `json:"val" bson:"val"`
	Set bool `json:"set" bson:"set"`
}

func OptFrom[T any](value T) Opt[T] {
	return Opt[T]{Val: value, Set: true}
}

func OptFromPtr[T any](ptr *T) Opt[T] {
	if ptr == nil {
		return Opt[T]{Set: false}
	}
	return Opt[T]{Val: *ptr, Set: true}
}
