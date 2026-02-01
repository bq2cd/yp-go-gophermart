package sqldatabase

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

// CreateOrder will insert a new record about order into the storage.
// It will return [false] if order with such ID already exists and
// an owner of the order.
// An error will be returned only in case of problems with the storage.
func (s *Storage) CreateOrder(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) (bool, domain.UserID, error) {
	user, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return false, domain.UserIDEmptyValue, err
	}

	if !exists {
		return false, domain.UserIDEmptyValue, domain.ErrUserNotFound
	}

	return s.maybeCreateOrder(ctx, user, orderID)
}

// GetOrders will return an array of orders for a given user ID.
// It will return an error in case of problems with the storage.
func (s *Storage) GetOrders(ctx context.Context, userID domain.UserID) ([]domain.Order, error) {
	user, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	orders, err := Query[models.Order](s).
		Where(generated.Order.UserID.Eq(user.ID)).
		Order(generated.Order.CreatedAt.Desc()).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot list orders: %w", err)
	}

	return makeDomainOrders(orders), nil
}

// GetOrderAccruals will return accrual points for provided order IDs.
// Accrual points might be stored/managed differently from the orders,
// and our domain model does not assume that an [domain.Order] has
// such a property.
// Hence, we provide a separate method to retrieve accrual points for
// the given orders.
func (s *Storage) GetOrderAccruals(
	ctx context.Context,
	userID domain.UserID,
	orderIDs []domain.OrderID,
) (map[domain.OrderID]float64, error) {
	_, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make(map[domain.OrderID]float64)

	if !exists {
		return result, nil
	}

	searchIDs := make([]uint, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		searchIDs = append(searchIDs, orderID.Uint())
	}

	accruals, err := Query[models.Accrual](s).
		Where(generated.Accrual.OrderID.In(searchIDs...)).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot search accruals: %w", err)
	}

	for _, accrual := range accruals {
		result[domain.OrderID(accrual.OrderID)] = accrual.Amount
	}

	return result, nil
}

// GetProcessableOrdersPerUser will return a mapping from a user ID to a list of order IDs
// belonging to that user and requiring further processing, that is, orders with statuses
// [domain.OrderStatusNew] and [domain.OrderStatusProcessing].
// The order IDs will be sorted from the oldest to the newest by [domain.Order.CreatedAt] field.
// This method is being used internally by [workers.OrderProcessor].
func (s *Storage) GetProcessableOrdersPerUser(
	ctx context.Context,
) (map[domain.UserID][]domain.OrderID, error) {
	result := make(map[domain.UserID][]domain.OrderID)

	searchValues := []int{
		domain.OrderStatusNew.Int(),
		domain.OrderStatusProcessing.Int(),
	}

	orders, err := Query[models.Order](s).
		Where(generated.Order.Status.In(searchValues...)).
		Preload("User", nil).
		Find(ctx)
	if err != nil {
		return result, fmt.Errorf("cannot search processable orders: %w", err)
	}

	for _, order := range orders {
		userID := domain.UserID(order.User.Login)
		result[userID] = append(result[userID], domain.OrderID(order.ID))
	}

	return result, nil
}

// GetOrderStatus will return order status as recorded in the storage.
// This method is being used internally by [workers.OrderProcessor].
// It will return [domain.ErrOrderNotFound] error if such order does not exist,
// and [domain.ErrUserIDConflict] if order belongs to a different user.
// Any other error will be returned in case of problems with the storage.
func (s *Storage) GetOrderStatus(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) (domain.OrderStatus, error) {
	order, err := s.findOrderBelongingToUser(ctx, userID, orderID)
	if err != nil {
		return domain.OrderStatusInvalid, err
	}

	return domain.OrderStatus(order.Status), nil
}

// MarkOrderInvalid will assign [domain.OrderStatusInvalid] status to
// a given order in the storage.
// It will return [domain.ErrOrderNotFound] error if such order does not exit,
// and [domain.ErrUserIDConflict] if order belongs to a different user.
// Any other error will be returned in case of problems with the storage.
func (s *Storage) MarkOrderInvalid(ctx context.Context, userID domain.UserID, orderID domain.OrderID) error {
	order, err := s.findOrderBelongingToUser(ctx, userID, orderID)
	if err != nil {
		return err
	}

	return s.setOrderStatus(ctx, order, domain.OrderStatusInvalid)
}

// MarkOrderProcessing will assign [domain.OrderStatusProcessing] status to
// a given order in the storage.
// It will return [domain.ErrOrderNotFound] error if such order does not exit,
// and [domain.ErrUserIDConflict] if order belongs to a different user.
// Any other error will be returned in case of problems with the storage.
func (s *Storage) MarkOrderProcessing(ctx context.Context, userID domain.UserID, orderID domain.OrderID) error {
	order, err := s.findOrderBelongingToUser(ctx, userID, orderID)
	if err != nil {
		return err
	}

	return s.setOrderStatus(ctx, order, domain.OrderStatusProcessing)
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
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	accrualPoints float64,
) error {
	return transactionMarkOrderProcessed(ctx, userID, orderID, accrualPoints).Run(s)
}

func (s *Storage) maybeCreateOrder(
	ctx context.Context,
	user models.User,
	orderID domain.OrderID,
) (bool, domain.UserID, error) {
	order, exists, err := s.findOrder(ctx, orderID)
	if err != nil {
		return false, domain.UserIDEmptyValue, err
	}

	if exists {
		return false, domain.UserID(order.User.Login), nil
	}

	//nolint:exhaustruct
	err = Query[models.Order](s).Create(ctx, &models.Order{
		ID:     orderID.Uint(),
		Status: domain.OrderStatusNew.Int(),
		UserID: user.ID,
	})
	if err != nil {
		return false, domain.UserIDEmptyValue, fmt.Errorf("cannot create order: %w", err)
	}

	return true, domain.UserID(user.Login), nil
}

func (s *Storage) findOrder(ctx context.Context, orderID domain.OrderID) (models.Order, bool, error) {
	order, err := Query[models.Order](s).
		Where(generated.Order.ID.Eq(orderID.Uint())).
		Preload("User", nil).
		First(ctx)

	switch {
	case err == nil:
		return order, true, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return order, false, nil
	default:
		return order, false, fmt.Errorf("cannot query orders: %w", err)
	}
}

func (s *Storage) findOrderBelongingToUser(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) (models.Order, error) {
	var order models.Order

	user, userExists, err := s.findUser(ctx, userID)
	if err != nil {
		return order, err
	}

	if !userExists {
		return order, domain.ErrUserNotFound
	}

	order, orderExists, err := s.findOrder(ctx, orderID)
	if err != nil {
		return order, err
	}

	if !orderExists || order.UserID != user.ID {
		return order, domain.ErrOrderNotFound
	}

	return order, nil
}

func (s *Storage) setOrderStatus(ctx context.Context, order models.Order, status domain.OrderStatus) error {
	_, err := Query[models.Order](s).
		Where(generated.Order.ID.Eq(order.ID)).
		Set(generated.Order.Status.Set(status.Int())).
		Update(ctx)
	if err != nil {
		return fmt.Errorf("cannot update order status: %w", err)
	}

	return nil
}

func (s *Storage) setAccrualPoints(ctx context.Context, order models.Order, amount float64) error {
	accrual, err := Query[models.Accrual](s).
		Where(generated.Accrual.OrderID.Eq(order.ID)).
		First(ctx)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		err = Query[models.Accrual](s).Create(ctx, &models.Accrual{
			OrderID: order.ID,
			Amount:  amount,
		})
	case err == nil:
		_, err = Query[models.Accrual](s).
			Where(generated.Accrual.OrderID.Eq(accrual.OrderID)).
			Set(generated.Accrual.Amount.Set(amount)).
			Update(ctx)
	default:
		return fmt.Errorf("cannot search accruals: %w", err)
	}

	if err != nil {
		return fmt.Errorf("cannot update accrual: %w", err)
	}

	return nil
}

// ConvertModelOrderToDomainOrder performs conversion of [models.Order]
// to [domain.Order] object.
func ConvertModelOrderToDomainOrder(modelOrder models.Order) domain.Order {
	return domain.Order{
		ID:        domain.OrderID(modelOrder.ID),
		Status:    domain.OrderStatus(modelOrder.Status),
		CreatedAt: modelOrder.CreatedAt,
	}
}

func makeDomainOrders(modelOrders []models.Order) []domain.Order {
	orders := make([]domain.Order, 0, len(modelOrders))

	for _, modelOrder := range modelOrders {
		orders = append(orders, ConvertModelOrderToDomainOrder(modelOrder))
	}

	return orders
}

func transactionMarkOrderProcessed(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	accrualPoints float64,
) Transaction {
	return Transaction{fn: func(stx *Storage) error {
		order, err := stx.findOrderBelongingToUser(ctx, userID, orderID)
		if err != nil {
			return err
		}

		err = stx.setOrderStatus(ctx, order, domain.OrderStatusProcessed)
		if err != nil {
			return err
		}

		err = stx.setAccrualPoints(ctx, order, accrualPoints)
		if err != nil {
			return err
		}

		err = stx.maybeUpdateBalance(ctx, order.User, accrualPoints, 0)
		if err != nil {
			return err
		}

		return nil
	}}
}
