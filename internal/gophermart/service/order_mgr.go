package service

import (
	"context"
	"fmt"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderManager implements [handler.OrderService] interface.
type OrderManager struct {
	orderRepo      OrderRepository
	orderProcessor OrderProcessor
}

// NewOrderManager creates an instance of [OrderManager].
func NewOrderManager(
	orderRepository OrderRepository,
	orderProcessor OrderProcessor,
) *OrderManager {
	return &OrderManager{
		orderRepo:      orderRepository,
		orderProcessor: orderProcessor,
	}
}

// CreateOrder performs creation of a new order
// and queues it for further processing.
func (m *OrderManager) CreateOrder(ctx context.Context, userID domain.UserID, orderID domain.OrderID) error {
	created, ownerID, err := m.orderRepo.CreateOrder(ctx, userID, orderID)
	if err != nil {
		return fmt.Errorf("cannot create order: %w", err)
	}

	if !created {
		return &domain.OrderIDAlreadyExistsError{
			OrderID:   orderID,
			CreatedBy: ownerID,
		}
	}

	m.orderProcessor.EnqueueOrder(ctx, userID, orderID)

	return nil
}

// GetOrders returns all orders by a given user
// sorted from the newest to the oldest.
func (m *OrderManager) GetOrders(ctx context.Context, userID domain.UserID) ([]domain.Order, error) {
	orders, err := m.orderRepo.GetOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve orders: %w", err)
	}

	domain.SortByTimestampFromNewestToOldest(orders)

	return orders, nil
}

// GetAccruals returns accrual points for provided order IDs.
func (m *OrderManager) GetAccruals(
	ctx context.Context,
	userID domain.UserID,
	orderIDs []domain.OrderID,
) (map[domain.OrderID]float64, error) {
	accruals, err := m.orderRepo.GetOrderAccruals(ctx, userID, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("cannot get order accrual points: %w", err)
	}

	return accruals, nil
}

// GetOrderIDs takes an array of orders and returns an array of order IDs.
func GetOrderIDs(orders []domain.Order) []domain.OrderID {
	ids := make([]domain.OrderID, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.ID)
	}

	return ids
}
