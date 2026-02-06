package sqldatabase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetNextProcessableOrder will return an order that is the best candidate to be processed immediately.
// It will also return a number of prior attempts of processing of this order.
// It will return an error if there are problems with the database.
func (s *Storage) GetNextProcessableOrder(
	ctx context.Context,
	excludeIDs []domain.OrderID,
) (domain.ProcessableOrder, uint, error) {
	var result domain.ProcessableOrder

	nextOrder, err := Query[models.ProcessableOrder](s, clauseLockForUpdateSkipLocked()).
		Joins(clause.RightJoin.Association("Order"), nil).
		Preload("Order.User", nil).
		Where(filterProcessableOrders(excludeIDs)).
		Order(generated.ProcessableOrder.ProcessAfter.Asc()).
		Take(ctx)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return result, 0, ErrNoProcessableOrders
	case err != nil:
		return result, 0, fmt.Errorf("cannot search processable orders: %w", err)
	}

	result = domain.ProcessableOrder{
		UserID:  domain.UserID(nextOrder.Order.User.Login),
		OrderID: domain.OrderID(nextOrder.Order.ID),
	}

	return result, nextOrder.Retries, nil
}

// PostponeOrderProcessing will postpone processing of a given order to a later date
// (as given by [processAfter] argument). This new date will be taken into account when [GetNextProcessableOrder]
// is called, so orders that have [processAfter] property in the future, will be excluded until that date comes.
func (s *Storage) PostponeOrderProcessing(
	ctx context.Context,
	order domain.ProcessableOrder,
	processAfter time.Time,
) error {
	return transactionPostponeOrderProcessing(ctx, order, processAfter).Run(s)
}

func (s *Storage) getProcessableOrder(ctx context.Context, orderID uint) (models.ProcessableOrder, error) {
	processable, err := Query[models.ProcessableOrder](s, clauseLockForUpdate()).
		Where(generated.ProcessableOrder.OrderID.Eq(orderID)).
		First(ctx)
	if err != nil {
		return processable, fmt.Errorf("cannot search processable orders: %w", err)
	}

	return processable, nil
}

func (s *Storage) removeProcessableOrder(ctx context.Context, order models.Order) error {
	processableOrder, err := s.getProcessableOrder(ctx, order.ID)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil
	case err != nil:
		return err
	}

	_, err = Query[models.ProcessableOrder](s).
		Where(generated.ProcessableOrder.OrderID.Eq(processableOrder.OrderID)).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("cannot remove processable order: %w", err)
	}

	return nil
}

//nolint:ireturn
func filterProcessableOrders(excludeIDs []domain.OrderID) clause.Expression {
	processableStatuses := []int{
		domain.OrderStatusNew.Int(),
		domain.OrderStatusProcessing.Int(),
	}

	excluded := make([]uint, 0, len(excludeIDs))
	for _, orderID := range excludeIDs {
		excluded = append(excluded, orderID.Uint())
	}

	return clause.And(
		generated.Order.Status.In(processableStatuses...),
		generated.Order.ID.NotIn(excluded...),
		generated.ProcessableOrder.ProcessAfter.Lte(time.Now()),
	)
}
