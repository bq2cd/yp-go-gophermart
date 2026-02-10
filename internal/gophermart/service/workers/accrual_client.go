package workers

import (
	"context"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
)

// AccrualClient is responsible for interaction with an accrual system.
// It provides various information about given order ID.
type AccrualClient interface {
	GetOrderStatus(ctx context.Context, orderID accdomain.OrderID) (accdomain.Order, error)
}
