package api

import (
	"time"
)

// OrderID represents the ID of an order.
type OrderID uint64

// OrderStatus defines possible values for an order status.
type OrderStatus string

const (
	// OrderStatusInvalid is used for invalid orders (e.g. not registered in the accrual system).
	// This is final status.
	OrderStatusInvalid OrderStatus = "INVALID"
	// OrderStatusNew is used for newly uploaded orders.
	OrderStatusNew OrderStatus = "NEW"
	// OrderStatusProcessed is used to mark order as processed.
	// This is final status.
	OrderStatusProcessed OrderStatus = "PROCESSED"
	// OrderStatusProcessing is used to mark an order that awaits further processing, e.g.
	// querying the accrual system, etc.
	OrderStatusProcessing OrderStatus = "PROCESSING"
)

// Order represents an order as uploaded by a user into the system.
type Order struct {
	Accrual    float64     `json:"accrual"`
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	UploadedAt time.Time   `json:"uploaded_at"` //nolint:tagliatelle
}

// Validate ensures that [OrderID] conforms to desired bounds.
func (oid OrderID) Validate() error {
	if oid == 0 {
		return ErrOrderIDMustBeGreaterThanZero
	}

	return nil
}
