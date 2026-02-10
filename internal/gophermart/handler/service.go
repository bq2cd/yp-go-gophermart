package handler

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// UserService provides methods to register/authenticate a user.
//
//go:generate mise run mockgen --outfile=user_service.go UserService
type UserService interface {
	Register(ctx context.Context, userID domain.UserID, passwordPlain domain.PasswordPlain) error
	Authenticate(ctx context.Context, userID domain.UserID, passwordPlain domain.PasswordPlain) error
}

// TokenService provides methods to issue and validate tokens for an authenticated user.
//
//go:generate mise run mockgen --outfile=token_service.go TokenService
type TokenService interface {
	IssueToken(ctx context.Context, userID domain.UserID) (domain.Token, error)
	ValidateToken(ctx context.Context, token domain.Token) (domain.UserID, error)
}

// BalanceService provides methods to get user's balance, perform a withdrawal or list prior withdrawals.
//
//go:generate mise run mockgen --outfile=balance_service.go BalanceService
type BalanceService interface {
	GetBalance(ctx context.Context, userID domain.UserID) (float64, error)
	GetTotalAmountWithdrawn(ctx context.Context, userID domain.UserID) (float64, error)
	GetWithdrawalTransactions(ctx context.Context, userID domain.UserID) ([]domain.WithdrawalTransaction, error)
	PayForOrderFromBalance(ctx context.Context, userID domain.UserID, orderID domain.OrderID, amount float64) error
}

// OrderService provides methods to upload new orders into the system
// and list prior user's orders.
//
//go:generate mise run mockgen --outfile=order_service.go OrderService
type OrderService interface {
	CreateOrder(ctx context.Context, userID domain.UserID, orderID domain.OrderID) error
	GetOrders(ctx context.Context, userID domain.UserID) ([]domain.Order, error)
	GetAccruals(
		ctx context.Context,
		userID domain.UserID,
		orderIDs []domain.OrderID,
	) (map[domain.OrderID]float64, error)
}
