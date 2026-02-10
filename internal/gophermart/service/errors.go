package service

import "errors"

var (
	// ErrTokenNotValid is returned by [TokenManager.ValidateToken] when parsed token
	// is not considered valid by [github.com/golang-jwt/jwt] package.
	ErrTokenNotValid = errors.New("JWT token is not valid")

	// ErrTokenInvalidUserID is returned by [TokenManager.ValidateToken] when parsed token
	// contains empty or otherwise malformed user ID.
	ErrTokenInvalidUserID = errors.New("JWT token contains invalid user ID")

	// ErrAmountIsNotPositive is returned by [BalanceManager.PayForOrderFromBalance] when
	// requested amount is not positive.
	ErrAmountIsNotPositive = errors.New("amount is not positive")

	// ErrOrderProcessorShuttingDown is returned by [OrderManager.CreateOrder] when
	// [OrderProcessor] is shutting down and cannot accept new orders.
	ErrOrderProcessorShuttingDown = errors.New("order processor is shutting down")
)
