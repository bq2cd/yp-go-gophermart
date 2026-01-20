package inmemory

import (
	"context"
	"fmt"
	"time"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// CreateOrder will insert a new record about order into the storage.
// It will return [false] if order with such ID already exists and
// an owner of the order.
// An error will be returned only in case of problems with the storage.
func (s *Storage) CreateOrder(
	_ context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) (bool, domain.UserID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if ok {
		return false, order.UserID, nil
	}

	order = Order{
		Order: domain.Order{
			ID:        orderID,
			Status:    domain.OrderStatusNew,
			CreatedAt: time.Now(),
		},
		UserID:        userID,
		AccrualPoints: 0,
	}

	s.orders[orderID] = order

	return true, userID, nil
}

// GetOrders will return an array of orders for a given user ID.
// It will return an error in case of problems with the storage.
func (s *Storage) GetOrders(_ context.Context, userID domain.UserID) ([]domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]domain.Order, 0)
	for _, order := range s.orders {
		if order.UserID != userID {
			continue
		}

		orders = append(orders, order.Order)
	}

	return orders, nil
}

// GetOrderAccruals will return accrual points for provided order IDs.
// Accrual points might be stored/managed differently from the orders,
// and our domain model does not assume that an [domain.Order] has
// such a property.
// Hence, we provide a separate method to retrieve accrual points for
// the given orders.
func (s *Storage) GetOrderAccruals(
	_ context.Context,
	userID domain.UserID,
	orderIDs []domain.OrderID,
) (map[domain.OrderID]float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accruals := map[domain.OrderID]float64{}
	for _, orderID := range orderIDs {
		order, err := s.getOrder(userID, orderID)
		if err != nil {
			continue
		}

		if order.AccrualPoints > 0 {
			accruals[orderID] = order.AccrualPoints
		}
	}

	return accruals, nil
}

// GetOrderStatus will return order status as recorded in the storage.
// This method is being used internally by [workers.OrderProcessor].
// It will return [domain.ErrOrderNotFound] error if such order does not exist,
// and [domain.ErrUserIDConflict] if order belongs to a different user.
// Any other error will be returned in case of problems with the storage.
func (s *Storage) GetOrderStatus(
	_ context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) (domain.OrderStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, err := s.getOrder(userID, orderID)
	if err != nil {
		return domain.OrderStatusInvalid, err
	}

	return order.Status, nil
}

// MarkOrderInvalid will assign [domain.OrderStatusInvalid] status to
// a given order in the storage.
// It will return [domain.ErrOrderNotFound] error if such order does not exit,
// and [domain.ErrUserIDConflict] if order belongs to a different user.
// Any other error will be returned in case of problems with the storage.
func (s *Storage) MarkOrderInvalid(_ context.Context, userID domain.UserID, orderID domain.OrderID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.setOrderStatusAndAccrualPoints(userID, orderID, domain.OrderStatusInvalid, 0)
}

// MarkOrderProcessing will assign [domain.OrderStatusProcessing] status to
// a given order in the storage.
// It will return [domain.ErrOrderNotFound] error if such order does not exit,
// and [domain.ErrUserIDConflict] if order belongs to a different user.
// Any other error will be returned in case of problems with the storage.
func (s *Storage) MarkOrderProcessing(_ context.Context, userID domain.UserID, orderID domain.OrderID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.setOrderStatusAndAccrualPoints(userID, orderID, domain.OrderStatusProcessing, 0)
}

// MarkOrderProcessed will assign [domain.OrderStatusProcessed] status to
// a given order in the storage,
// assign provided accrual points to the order,
// and update user's balance.
// It will ensure atomic update of the user's balance to prevent
// race conditions with withdrawal operations.
// It will return [domain.ErrOrderNotFound] error if such order does not exit,
// and [domain.ErrUserIDConflict] if order belongs to a different user.
// Any other error will be returned in case of problems with the storage.
func (s *Storage) MarkOrderProcessed(
	_ context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	accrualPoints float64,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.setOrderStatusAndAccrualPoints(userID, orderID, domain.OrderStatusProcessed, accrualPoints)
	if err != nil {
		return err
	}

	success, err := s.updateUser(userID, func(u *User) (bool, error) {
		return u.addFunds(accrualPoints)
	})
	if err != nil {
		return fmt.Errorf("cannot update user's balance: %w", err)
	}

	if !success {
		return ErrUserBalanceCannotBeUpdated
	}

	return nil
}

func (s *Storage) getOrder(userID domain.UserID, orderID domain.OrderID) (Order, error) {
	order, ok := s.orders[orderID]
	if !ok {
		return Order{}, domain.ErrOrderNotFound
	}

	if order.UserID != userID {
		return Order{}, domain.ErrUserIDConflict
	}

	return order, nil
}

// setOrderStatusAndAccrualPoints is intended to be executed atomically.
// It is the responsibility of a caller to ensure that this call is
// protected from race conditions.
func (s *Storage) setOrderStatusAndAccrualPoints(
	userID domain.UserID,
	orderID domain.OrderID,
	status domain.OrderStatus,
	accrualPoints float64,
) error {
	order, err := s.getOrder(userID, orderID)
	if err != nil {
		return err
	}

	//nolint: exhaustive
	switch order.Status {
	case domain.OrderStatusInvalid, domain.OrderStatusProcessed:
		return ErrOrderStatusIsFinal
	}

	order.Status = status
	order.AccrualPoints = accrualPoints

	s.orders[orderID] = order

	return nil
}
