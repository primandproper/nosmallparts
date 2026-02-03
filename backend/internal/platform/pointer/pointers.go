package pointer

// To returns a pointer to a value.
func To[T any](x T) *T {
	return &x
}

// Dereference returns the value of a pointer.
func Dereference[T any](x *T) T {
	if x == nil {
		var zero T
		return zero
	}
	return *x
}

// DereferenceSlice returns the value of a pointer for every element in a slice.
func DereferenceSlice[T any](x []*T) []T {
	if x == nil {
		return []T{}
	}

	y := make([]T, len(x))
	for i := range x {
		y[i] = *x[i]
	}
	return y
}
