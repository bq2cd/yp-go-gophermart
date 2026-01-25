package app

import "errors"

var (
	// ErrEmptyAccrualSystemAddress is returned when supplied accrual system address is empty.
	ErrEmptyAccrualSystemAddress = errors.New("accrual system address cannot be empty")

	// ErrEmptyListenAddress is returned when HTTP server listen address is empty.
	ErrEmptyListenAddress = errors.New("listen address cannot be empty")
)
