package domain

import "strconv"

// OrderID represents the ID of an [Order].
type OrderID uint64

// String renders order ID as string.
func (oid OrderID) String() string {
	return strconv.FormatUint(uint64(oid), 10)
}

// OrderStatus represent the status of an [Order].
type OrderStatus int

const (
	// OrderStatusInvalid is assigned to an [Order] when accrual system cannot process it.
	// This is a final status.
	OrderStatusInvalid OrderStatus = iota - 1
	// OrderStatusRegistered is assigned to an [Order] when accrual system registers a new order.
	OrderStatusRegistered
	// OrderStatusProcessing is assigned to an [Order] when accrual system starts processing the order.
	OrderStatusProcessing
	// OrderStatusProcessed is assigned to an [Order] when accrual system finished processing of the order
	// and assigned accrual points to the order.
	// This is a final status.
	OrderStatusProcessed
)

// Order combines essential information about a user's order in the accrual system.
type Order struct {
	ID            OrderID
	Status        OrderStatus
	AccrualPoints float64
}
