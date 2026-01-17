package service

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderProcessor is responsible for background processing
// of newly created orders.
//
//go:generate go tool mockgen -typed -destination=mocks/order_processor.go -package=mocks . OrderProcessor
type OrderProcessor interface {
	// EnqueueOrder places order ID into in-memory queue to perform further processing.
	EnqueueOrder(userID domain.UserID, orderID domain.OrderID) bool
}
