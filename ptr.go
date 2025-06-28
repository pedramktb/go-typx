package typx

// Ptr for inline usage (e.g. output of a function)
func Ptr[T any](value T) *T {
	return &value
}

// Safe pointer deference
func FromPtr[T any](pointer *T) (T, bool) {
	if pointer == nil {
		return *new(T), false
	}
	return *pointer, true
}
