package sqldatabase

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

func transactionWithdrawFunds(
	ctx context.Context,
	user models.User,
	orderID domain.OrderID,
	amount float64,
) Transaction {
	return Transaction{fn: func(stx *Storage) error {
		err := stx.maybeUpdateBalance(ctx, user, -1*amount, amount)
		if err != nil {
			return err
		}

		return stx.addWithdrawalTransaction(ctx, user, orderID, amount)
	}}
}
