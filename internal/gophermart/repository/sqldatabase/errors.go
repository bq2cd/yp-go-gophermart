package sqldatabase

import "errors"

var (
	// ErrUnsupportedDatabaseDriver is returned by [NewStorage] when it encounters a database driver
	// that is not supported.
	ErrUnsupportedDatabaseDriver = errors.New("unsupported database driver")

	// ErrWithdrawalAmountMustBePositive is returned by [Storage.WithdrawFunds] method when requested withdrawal
	// amount is below zero.
	ErrWithdrawalAmountMustBePositive = errors.New("withdrawal amount must be positive")
)
