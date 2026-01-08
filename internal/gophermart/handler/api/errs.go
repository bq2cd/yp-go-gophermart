package api

import "errors"

var (
	// ErrOrderIDMustBeGreaterThanZero is returned during validation phase when order ID is zero.
	ErrOrderIDMustBeGreaterThanZero = errors.New("order ID must be greater than zero")
)
