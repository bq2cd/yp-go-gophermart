package workers

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderRepository is responsible for updating order status and user's balance,
// as well as providing a status of a given order.
type OrderRepository interface {
	GetOrderStatus(ctx context.Context, userID domain.UserID, orderID domain.OrderID) (domain.OrderStatus, error)
	MarkOrderInvalid(ctx context.Context, userID domain.UserID, orderID domain.OrderID) error
	MarkOrderProcessing(ctx context.Context, userID domain.UserID, orderID domain.OrderID) error
	// MarkOrderProcessed is responsible for setting order status
	// as [domain.OrderStatusProcessed] and updating user's balance
	// with provided accrual points.
	// It should perform all updates atomically.
	MarkOrderProcessed(ctx context.Context, userID domain.UserID, orderID domain.OrderID, amount float64) error
}
