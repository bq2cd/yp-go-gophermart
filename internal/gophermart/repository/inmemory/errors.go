package inmemory

import "errors"

var (
	// ErrOrderStatusIsFinal is returned when there's an attempt to change
	// order status from final status.
	// Final statuses are [domain.OrderStatusInvalid] and [domain.OrderStatusProcessed].
	ErrOrderStatusIsFinal = errors.New("order final status cannot be changed")

	// ErrUserBalanceCannotBeUpdated is returned when [User.addFunds] would return [false],
	// which should rarely be a case - adding to user's balance should succeed
	// unless there has been an error in arithmetic.
	ErrUserBalanceCannotBeUpdated = errors.New("user's balance cannot be updated")
)
