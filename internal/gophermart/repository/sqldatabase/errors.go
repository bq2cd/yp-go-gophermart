package sqldatabase

import "errors"

var (
	// ErrUnsupportedDatabaseDriver is returned by [NewStorage] when it encounters a database driver
	// that is not supported.
	ErrUnsupportedDatabaseDriver = errors.New("unsupported database driver")

	// ErrNoProcessableOrders is returned by [Storage.GetNextProcessableOrder] when there are no orders
	// available for processing.
	ErrNoProcessableOrders = errors.New("no processable orders")
)
