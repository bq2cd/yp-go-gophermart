package service

import (
	"context"
	"fmt"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// BalanceManager implements [handler.BalanceService] interface.
type BalanceManager struct {
	balanceRepo BalanceRepository
}

// NewBalanceManager creates an instance of [BalanceManager].
func NewBalanceManager(balanceRepository BalanceRepository) *BalanceManager {
	return &BalanceManager{
		balanceRepo: balanceRepository,
	}
}

// GetBalance returns current value of user's accumulated accrual points.
func (m *BalanceManager) GetBalance(ctx context.Context, userID domain.UserID) (float64, error) {
	balance, err := m.balanceRepo.GetCurrentValue(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("cannot retrieve balance: %w", err)
	}

	return balance, nil
}

// GetTotalAmountWithdrawn returns cumulative value of all withdrawal transactions made by a user.
func (m *BalanceManager) GetTotalAmountWithdrawn(ctx context.Context, userID domain.UserID) (float64, error) {
	total, err := m.balanceRepo.GetTotalAmountWithdrawn(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("cannot retrieve total withdrawals: %w", err)
	}

	return total, nil
}

// GetWithdrawalTransactions returns an array of withdrawal transactions made by a user,
// sorted from the newest to the oldest.
func (m *BalanceManager) GetWithdrawalTransactions(
	ctx context.Context,
	userID domain.UserID,
) ([]domain.WithdrawalTransaction, error) {
	transactions, err := m.balanceRepo.GetWithdrawalTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve withdrawal transactions: %w", err)
	}

	domain.SortByTimestampFromNewestToOldest(transactions)

	return transactions, nil
}

// PayForOrderFromBalance performs a withdrawal of a specified amount
// of accrual points from user's balance to cover a part of the cost
// of a given order.
func (m *BalanceManager) PayForOrderFromBalance(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	amount float64,
) error {
	if amount <= 0 {
		return ErrAmountIsNotPositive
	}

	success, err := m.balanceRepo.WithdrawFunds(ctx, userID, orderID, amount)
	if err != nil {
		return fmt.Errorf("cannot withdraw funds: %w", err)
	}

	if !success {
		return domain.ErrBalanceNotEnoughFunds
	}

	return nil
}
