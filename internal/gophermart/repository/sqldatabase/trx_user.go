package sqldatabase

import (
	"context"
	"fmt"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

func transactionCreateUserAndBalance(
	ctx context.Context,
	userID domain.UserID,
	passwordHash domain.PasswordHash,
) Transaction {
	return Transaction{fn: func(stx *Storage) error {
		//nolint:exhaustruct
		user := &models.User{
			Login:        userID.String(),
			PasswordHash: passwordHash.Bytes(),
		}

		err := Query[models.User](stx).Create(ctx, user)
		if err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}

		balance := &models.Balance{
			UserID:    user.ID,
			Current:   0,
			Withdrawn: 0,
		}

		err = Query[models.Balance](stx).Create(ctx, balance)
		if err != nil {
			return fmt.Errorf("cannot create balance: %w", err)
		}

		return nil
	}}
}
