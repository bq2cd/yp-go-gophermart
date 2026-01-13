package api

import "errors"

var (
	// ErrOrderIDMustBeGreaterThanZero is returned during validation phase when order ID is zero.
	ErrOrderIDMustBeGreaterThanZero = errors.New("order ID must be greater than zero")

	// ErrOrderIDLuhnChecksumMismatch is returned during validation phase when
	// calculated Luhn's checksum is invalid, which means that order ID contains wrong digits.
	ErrOrderIDLuhnChecksumMismatch = errors.New("order ID has invalid Luhn's checksum")
)
