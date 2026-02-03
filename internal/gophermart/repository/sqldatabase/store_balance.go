package sqldatabase

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

// WithdrawFunds will attempt to withdraw requested amount from user's balance.
// It will return [false] if there are not enough funds on the balance.
// For non-existent users, it will return [domain.ErrUserNotFound] error.
// Other errors will be returned in case of problems with the storage.
func (s *Storage) WithdrawFunds(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	amount float64,
) (bool, error) {
	user, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, domain.ErrUserNotFound
	}

	err = transactionWithdrawFunds(ctx, user, orderID, amount).Run(s)
	if err != nil {
		return false, fmt.Errorf("cannot withdraw funds: %w", err)
	}

	return true, nil
}

// GetCurrentValue will return current value of user's balance.
// For non-existent users, it will return [domain.ErrUserNotFound] error.
// Other errors will be returned in case of problems with the storage.
func (s *Storage) GetCurrentValue(ctx context.Context, userID domain.UserID) (float64, error) {
	balance, err := s.findBalance(ctx, userID)
	if err != nil {
		return 0, err
	}

	return balance.Current, nil
}

// GetTotalAmountWithdrawn will return total amount of all user's withdrawals,
// that is, a sum of [domain.WithdrawalTransaction] transactions.
// For non-existent users, it will return [domain.ErrUserNotFound] error.
// Other errors will be returned in case of problems with the storage.
// The reason this method is separate is that this sum can be
// efficiently updated internally on every [WithdrawFunds] call,
// thus avoiding potentially expensive calculation from the array
// of transactions.
func (s *Storage) GetTotalAmountWithdrawn(ctx context.Context, userID domain.UserID) (float64, error) {
	balance, err := s.findBalance(ctx, userID)
	if err != nil {
		return 0, err
	}

	return balance.Withdrawn, nil
}

// GetWithdrawalTransactions will return all user's withdrawal transaction.
// It will return an error in case of problems with the storage.
func (s *Storage) GetWithdrawalTransactions(
	ctx context.Context,
	userID domain.UserID,
) ([]domain.WithdrawalTransaction, error) {
	user, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	withdrawals, err := Query[models.Withdrawal](s).
		Where(generated.Withdrawal.UserID.Eq(user.ID)).
		Order(generated.Withdrawal.CreatedAt.Desc()).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot search withdrawals: %w", err)
	}

	transactions := make([]domain.WithdrawalTransaction, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		transactions = append(transactions, domain.WithdrawalTransaction{
			OrderID:     domain.OrderID(withdrawal.NextOrderID),
			Amount:      withdrawal.Amount,
			ProcessedAt: withdrawal.CreatedAt,
		})
	}

	return transactions, nil
}

func (s *Storage) findBalance(ctx context.Context, userID domain.UserID) (models.Balance, error) {
	var balance models.Balance

	user, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return balance, err
	}

	if !exists {
		return balance, nil
	}

	return s.getBalance(ctx, user)
}

func (s *Storage) getBalance(ctx context.Context, user models.User, opts ...clause.Expression) (models.Balance, error) {
	balance, err := Query[models.Balance](s, opts...).
		Where(generated.Balance.UserID.Eq(user.ID)).
		First(ctx)
	if err != nil {
		return balance, fmt.Errorf("cannot search balances: %w", err)
	}

	return balance, nil
}

// maybeUpdateBalance is intended to be run inside a transaction,
// since it relies on the transaction to be rolled back if
// this function returns an error.
func (s *Storage) maybeUpdateBalance(ctx context.Context, user models.User, diffCurrent, diffWithdrawn float64) error {
	err := s.updateBalance(ctx, user, diffCurrent, diffWithdrawn)
	if err != nil {
		return err
	}

	balance, err := s.getBalance(ctx, user)
	if err != nil {
		return err
	}

	if balance.Current < 0 {
		return domain.ErrBalanceNotEnoughFunds
	}

	return nil
}

func (s *Storage) updateBalance(ctx context.Context, user models.User, diffCurrent, diffWithdrawn float64) error {
	//nolint:exhaustruct
	lockForUpdate := clause.Locking{
		Strength: clause.LockingStrengthUpdate,
	}

	balance, err := s.getBalance(ctx, user, lockForUpdate)
	if err != nil {
		return err
	}

	// Coarse-level check if balance can go negative.
	// If integer part of the balance remains positive,
	// then we defer to the database to do the rounding of the
	// fractional part by performing an update and
	// checking its result.
	if int(balance.Current)+int(diffCurrent) < 0 {
		return domain.ErrBalanceNotEnoughFunds
	}

	_, err = Query[models.Balance](s).
		Where(generated.Balance.UserID.Eq(user.ID)).
		Set(
			generated.Balance.Current.Incr(diffCurrent),
			generated.Balance.Withdrawn.Incr(diffWithdrawn),
		).
		Update(ctx)
	if err != nil {
		return fmt.Errorf("cannot update balance: %w", err)
	}

	return nil
}

func (s *Storage) addWithdrawalTransaction(
	ctx context.Context,
	user models.User,
	orderID domain.OrderID,
	amount float64,
) error {
	//nolint:exhaustruct
	withdrawal := &models.Withdrawal{
		Amount:      amount,
		NextOrderID: orderID.Uint(),
		UserID:      user.ID,
	}

	err := Query[models.Withdrawal](s).Create(ctx, withdrawal)
	if err != nil {
		return fmt.Errorf("cannot add withdrawal transaction: %w", err)
	}

	return nil
}

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
