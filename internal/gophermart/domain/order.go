package domain

import "time"

// OrderID represents the ID of an [Order].
type OrderID uint64

// OrderStatus represent the status of an [Order].
type OrderStatus int

const (
	// OrderStatusInvalid is assigned to an [Order] when it fails business validation.
	// This is a final status.
	OrderStatusInvalid OrderStatus = iota - 1
	// OrderStatusNew is assigned to newly created [Order].
	OrderStatusNew
	// OrderStatusProcessing is assigned to an [Order] that has been put into an internal processing pipeline.
	OrderStatusProcessing
	// OrderStatusProcessed is assigned to an [Order] when it has been fully processed.
	// This is a final status.
	OrderStatusProcessed
)

// Order combines essential information about a user's order in the system.
type Order struct {
	ID        OrderID
	Status    OrderStatus
	CreatedAt time.Time
}
