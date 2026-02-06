package sqldatabase

import (
	"context"
	"fmt"
	"time"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

func transactionCreateOrder(
	ctx context.Context,
	user models.User,
	orderID domain.OrderID,
) Transaction {
	return Transaction{fn: func(stx *Storage) error {
		//nolint:exhaustruct
		err := Query[models.Order](stx).Create(ctx, &models.Order{
			ID:     orderID.Uint(),
			Status: domain.OrderStatusNew.Int(),
			UserID: user.ID,
		})
		if err != nil {
			return fmt.Errorf("cannot create order: %w", err)
		}

		//nolint:exhaustruct
		err = Query[models.ProcessableOrder](stx).Create(ctx, &models.ProcessableOrder{
			OrderID:      orderID.Uint(),
			ProcessAfter: time.Now(),
			Retries:      0,
		})
		if err != nil {
			return fmt.Errorf("cannot create processable order: %w", err)
		}

		return nil
	}}
}

func transactionMarkOrderInvalid(
	ctx context.Context,
	order models.Order,
) Transaction {
	return Transaction{fn: func(stx *Storage) error {
		err := stx.setOrderStatus(ctx, order, domain.OrderStatusInvalid)
		if err != nil {
			return err
		}

		return stx.removeProcessableOrder(ctx, order)
	}}
}

func transactionMarkOrderProcessed(
	ctx context.Context,
	order models.Order,
	accrualPoints float64,
) Transaction {
	return Transaction{fn: func(stx *Storage) error {
		err := stx.setOrderStatus(ctx, order, domain.OrderStatusProcessed)
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

		return stx.removeProcessableOrder(ctx, order)
	}}
}
