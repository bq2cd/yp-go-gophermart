package inmemory

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// WithdrawFunds will attempt to withdraw requested amount from user's balance.
// It will return [false] if there are not enough funds on the balance.
// For non-existent users, it will return [domain.ErrUserNotFound] error.
// Other errors will be returned in case of problems with the storage.
func (s *Storage) WithdrawFunds(
	_ context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	amount float64,
) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var success bool

	success, err := s.updateUser(userID, func(u *User) (bool, error) {
		return u.withdrawFunds(amount)
	})
	if err != nil {
		return false, fmt.Errorf("cannot withdraw funds: %w", err)
	}

	if !success {
		return false, nil
	}

	s.addWithdrawalTransaction(userID, orderID, amount)

	return true, nil
}

// GetCurrentValue will return current value of user's balance.
// For non-existent users, it will return [domain.ErrUserNotFound] error.
// Other errors will be returned in case of problems with the storage.
func (s *Storage) GetCurrentValue(_ context.Context, userID domain.UserID) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return 0, domain.ErrUserNotFound
	}

	return user.GetBalanceAsFloat(), nil
}

// GetTotalAmountWithdrawn will return total amount of all user's withdrawals,
// that is, a sum of [domain.WithdrawalTransaction] transactions.
// For non-existent users, it will return [domain.ErrUserNotFound] error.
// Other errors will be returned in case of problems with the storage.
// The reason this method is separate is that this sum can be
// efficiently updated internally on every [WithdrawFunds] call,
// thus avoiding potentially expensive calculation from the array
// of transactions.
func (s *Storage) GetTotalAmountWithdrawn(_ context.Context, userID domain.UserID) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return 0, domain.ErrUserNotFound
	}

	return user.GetTotalAmountWithdrawnAsFloat(), nil
}

// GetWithdrawalTransactions will return all user's withdrawal transaction.
// It will return an error in case of problems with the storage.
func (s *Storage) GetWithdrawalTransactions(
	_ context.Context,
	userID domain.UserID,
) ([]domain.WithdrawalTransaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	withdrawals, ok := s.withdrawals[userID]
	if !ok {
		return nil, nil
	}

	sorted := slices.Clone(withdrawals)

	domain.SortByTimestampFromNewestToOldest(sorted)

	return sorted, nil
}

// addWithdrawalTransaction is intended to be executed atomically.
// It is the responsibility of a caller to ensure that this call is
// protected from race conditions.
func (s *Storage) addWithdrawalTransaction(userID domain.UserID, orderID domain.OrderID, amount float64) {
	trx := domain.WithdrawalTransaction{
		OrderID:     orderID,
		Amount:      amount,
		ProcessedAt: time.Now(),
	}

	s.withdrawals[userID] = append(s.withdrawals[userID], trx)
}
