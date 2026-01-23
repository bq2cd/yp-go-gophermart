package testutil

import "math/rand/v2"

// RandomSliceElement returns a pseudo-random element from a given slice
// or a zero value if the slice was empty.
//
//nolint:ireturn
func RandomSliceElement[T any](input []T) T {
	var zero T

	if len(input) == 0 {
		return zero
	}

	idx := rand.IntN(len(input)) //nolint:gosec

	return input[idx]
}
