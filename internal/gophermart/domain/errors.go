package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrUserIDConflict is returned during user registration process if a user already
	// exists with given [UserID].
	ErrUserIDConflict = errors.New("user with such ID already exists")

	// ErrUserAuthenticationFailed is returned during user authentication process when
	// there is either a credentials mismatch or user does not exist.
	ErrUserAuthenticationFailed = errors.New("user authentication failed")

	// ErrBalanceNotEnoughFunds is returned when user attempts to pay for an order with funds from the balance,
	// but there are not enough funds for that.
	ErrBalanceNotEnoughFunds = errors.New("not enough funds on balance")

	// ErrOrderIDValidationFailed is returned whenever incoming order ID is validated on the service level.
	// This might involve checksum calculation or order ID lookup in external systems.
	ErrOrderIDValidationFailed = errors.New("invalid order ID")
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
