package typx

// [Must] is a helper that panics if the error is non-nil, otherwise returning the value.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// [Must2] is a helper that panics if the error is non-nil, otherwise returning the two values.
func Must2[T, U any](val T, val2 U, err error) (T, U) {
	if err != nil {
		panic(err)
	}
	return val, val2
}

// [Must3] is a helper that panics if the error is non-nil, otherwise returning the three values.
func Must3[T, U, V any](val T, val2 U, val3 V, err error) (T, U, V) {
	if err != nil {
		panic(err)
	}
	return val, val2, val3
}
