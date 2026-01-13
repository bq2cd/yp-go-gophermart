package luhn

import "errors"

var (
	// ErrCheckDigitMismatch is returned from [Validate] when Luhn's calculated check digit
	// does not match the last digit of a given number.
	ErrCheckDigitMismatch = errors.New("check digit mismatch")
)
