package typx

// [Ptr] creates a pointer to the receiver copy of the given value.
// Useful for inline pointer creation such as function calls.
//
// Deprecated: Since Go 1.26, you can use the built-in new(value) instead.
func Ptr[T any](value T) *T {
	return &value
}

// [FromPtr] safely dereferences the given pointer and returns it along with a boolean
// indicating whether the pointer was non-nil.
func FromPtr[T any](ptr *T) (T, bool) {
	if ptr == nil {
		return *new(T), false
	}
	return *ptr, true
}

// [FromPtrOrZero] safely dereferences the given pointer and returns its value.
// If the pointer is nil, it returns the zero value of T.
func FromPtrOrZero[T any](ptr *T) T {
	if ptr == nil {
		return *new(T)
	}
	return *ptr
}
