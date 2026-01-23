package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrUserNotFound is returned whenever we check if user with given [UserID]
	// exists in the system and fail to find such a user.
	ErrUserNotFound = errors.New("user with such ID does not exist")

	// ErrUserIDConflict is returned during user registration process if a user already
	// exists with given [UserID].
	ErrUserIDConflict = errors.New("user with such ID already exists")

	// ErrUserAuthenticationFailed is returned during user authentication process when
	// there is either a credentials mismatch or user does not exist.
	ErrUserAuthenticationFailed = errors.New("user authentication failed")

	// ErrEmptyPassword is returned during user registration process when [PasswordPlain]
	// is an empty string.
	ErrEmptyPassword = errors.New("password is empty")

	// ErrBalanceNotEnoughFunds is returned when user attempts to pay for an order with funds from the balance,
	// but there are not enough funds for that.
	ErrBalanceNotEnoughFunds = errors.New("not enough funds on balance")

	// ErrOrderNotFound is returned when an order with a given ID does not exist in the system.
	ErrOrderNotFound = errors.New("order with such ID does not exist")
)

// OrderIDAlreadyExistsError is returned whenever a user attempts to create a new order,
// but another order with such ID already exists in the system.
type OrderIDAlreadyExistsError struct {
	OrderID   OrderID
	CreatedBy UserID
}

func (e *OrderIDAlreadyExistsError) Error() string {
	return fmt.Sprintf("order ID %d already exists (created by %s)", e.OrderID, e.CreatedBy)
}
