package workers

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderEvent represent an internal event for the [OrderProcessor]
// used during processing.
type OrderEvent struct {
	userID  domain.UserID
	orderID domain.OrderID
}

type orderEventResult struct {
	OrderEvent

	err error
}
