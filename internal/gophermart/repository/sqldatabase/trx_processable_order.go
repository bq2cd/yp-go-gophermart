package sqldatabase

import (
	"context"
	"fmt"
	"time"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
	"gorm.io/gorm/clause"
)

func transactionPostponeOrderProcessing(
	ctx context.Context,
	order domain.ProcessableOrder,
	processAfter time.Time,
) Transaction {
	return Transaction{fn: func(stx *Storage) error {
		processable, err := stx.getProcessableOrder(ctx, order.OrderID.Uint())
		if err != nil {
			return err
		}

		processable.ProcessAfter = processAfter
		processable.Retries++

		_, err = Query[models.ProcessableOrder](stx).
			Omit(clause.Associations).
			Updates(ctx, processable)
		if err != nil {
			return fmt.Errorf("cannot update processable order: %w", err)
		}

		return nil
	}}
}
