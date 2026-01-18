package service

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderRepository is responsible for creating a new order and retrieving information about user's orders.
//
//go:generate go tool mockgen -typed -destination=mocks/order_repository.go -package=mocks . OrderRepository
type OrderRepository interface {
	CreateOrder(ctx context.Context, userID domain.UserID, orderID domain.OrderID) (bool, domain.UserID, error)
	GetOrders(ctx context.Context, userID domain.UserID) ([]domain.Order, error)
	GetOrderAccruals(
		ctx context.Context,
		userID domain.UserID,
		orderIDs []domain.OrderID,
	) (map[domain.OrderID]float64, error)
}
