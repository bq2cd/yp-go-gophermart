package service

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// BalanceRepository is responsible for adding/withdrawing funds
//
// to user's balance, as well as retrieving current balance value
// and a list of withdrawal transactions.
//
//go:generate go tool mockgen -typed -destination=mocks/balance_repository.go -package=mocks . BalanceRepository
type BalanceRepository interface {
	WithdrawFunds(ctx context.Context, userID domain.UserID, amount float64) (bool, error)
	GetCurrentValue(ctx context.Context, userID domain.UserID) (float64, error)
	GetTotalAmountWithdrawn(ctx context.Context, userID domain.UserID) (float64, error)
	GetWithdrawalTransactions(ctx context.Context, userID domain.UserID) ([]domain.WithdrawalTransaction, error)
}
