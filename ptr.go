package typx

// Ptr creates a pointer to the receiver copy of the given value.
// Useful for inline pointer creation such as function calls.
func Ptr[T any](value T) *T {
	return &value
}

// FromPtr safely dereferences the given pointer and returns it along with a boolean
// indicating whether the pointer was non-nil.
func FromPtr[T any](pointer *T) (T, bool) {
	if pointer == nil {
		return *new(T), false
	}
	return *pointer, true
}
