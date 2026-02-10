package service

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"

// OrderProcessor is responsible for background processing
// of newly created orders.
//
//go:generate mise run mockgen --outfile=order_processor.go OrderProcessor
type OrderProcessor interface {
	// NewOrderArrived notifies order processor about newly created order.
	// It is up to the order processor when and if to start the processing.
	NewOrderArrived(userID domain.UserID, orderID domain.OrderID)
}
