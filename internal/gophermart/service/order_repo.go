package service

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderRepository is responsible for storing/updating/retrieving information
//
//	about orders.
//
//go:generate go tool mockgen -typed -destination=mocks/order_repository.go -package=mocks . OrderRepository
type OrderRepository interface {
	CreateOrder(ctx context.Context, userID domain.UserID, orderID domain.OrderID) (bool, domain.UserID, error)
	SetOrderStatus(ctx context.Context, userID domain.UserID, orderID domain.OrderID, status domain.OrderStatus) error
	SetOrderAccrual(ctx context.Context, userID domain.UserID, orderID domain.OrderID, amount float64) error
	GetOrders(ctx context.Context, userID domain.UserID) ([]domain.Order, error)
	GetOrderAccruals(
		ctx context.Context,
		userID domain.UserID,
		orderIDs []domain.OrderID,
	) (map[domain.OrderID]float64, error)
}
