package inmemory

import (
	"fmt"

	"github.com/govalues/decimal"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

const (
	decimalDigitsAfterDecimalPoint = 4
)

// UserUpdateFn defines a function that will perform an update of the [User] data.
type UserUpdateFn func(*User) (bool, error)

// User combines user password, balance, and total amount of withdrawals.
// This is intended to be used for internal storage.
type User struct {
	Password             domain.PasswordHash
	Balance              decimal.Decimal
	TotalAmountWithdrawn decimal.Decimal
}

// GetBalanceAsFloat converts internal decimal representation of
// user's balance to [float64].
func (u *User) GetBalanceAsFloat() float64 {
	value, _ := u.Balance.Float64()

	return value
}

// GetTotalAmountWithdrawnAsFloat converts internal decimal representation of
// user's total withdrawals' amount to [float64].
func (u *User) GetTotalAmountWithdrawnAsFloat() float64 {
	value, _ := u.TotalAmountWithdrawn.Float64()

	return value
}

// withdrawFunds is intended to be executed atomically.
// It is the responsibility of a caller to ensure that this call is
// protected from race conditions.
// The result will be [false] if there are not enough funds on the
// balance.
// An error will be returned if there was a loss of minimal precision
// during arithmetic operations.
func (u *User) withdrawFunds(amount float64) (bool, error) {
	outgoing, err := decimal.NewFromFloat64(amount)
	if err != nil {
		return false, fmt.Errorf("cannot convert amount to decimal: %w", err)
	}

	newBalance, err := u.Balance.SubExact(outgoing, decimalDigitsAfterDecimalPoint)
	if err != nil {
		return false, fmt.Errorf("cannot subtract decimal amount exactly: %w", err)
	}

	if newBalance.Cmp(decimal.Zero) < 0 {
		return false, nil
	}

	newAmountWithdrawn, err := u.TotalAmountWithdrawn.AddExact(outgoing, decimalDigitsAfterDecimalPoint)
	if err != nil {
		return false, fmt.Errorf("cannot add decimal amount exactly: %w", err)
	}

	u.Balance = newBalance
	u.TotalAmountWithdrawn = newAmountWithdrawn

	return true, nil
}

// addFunds is intended to be executed atomically.
// It is the responsibility of a caller to ensure that this call is
// protected from race conditions.
// The result will be [true] if no errors occurred; this is to conform to [UserUpdateFn] signature.
// An error will be returned if there was a loss of minimal precision
// during arithmetic operations.
func (u *User) addFunds(amount float64) (bool, error) {
	incoming, err := decimal.NewFromFloat64(amount)
	if err != nil {
		return false, fmt.Errorf("cannot convert amount to decimal: %w", err)
	}

	newBalance, err := u.Balance.AddExact(incoming, decimalDigitsAfterDecimalPoint)
	if err != nil {
		return false, fmt.Errorf("cannot add decimal amount exactly: %w", err)
	}

	u.Balance = newBalance

	return true, nil
}
