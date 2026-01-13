package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"

// UserService provides methods to register/authenticate a user.
//
//go:generate go tool mockgen -typed -destination=mocks/user_service.go -package=mocks . UserService
type UserService interface {
	Register(userID domain.UserID, passwordPlain domain.PasswordPlain) error
	Authenticate(userID domain.UserID, passwordPlain domain.PasswordPlain) error
}

// TokenService provides methods to issue and validate tokens for an authenticated user.
//
//go:generate go tool mockgen -typed -destination=mocks/token_service.go -package=mocks . TokenService
type TokenService interface {
	IssueToken(userID domain.UserID) (domain.Token, error)
	ValidateToken(token domain.Token) (domain.UserID, error)
}

// BalanceService provides methods to get user's balance, perform a withdrawal or list prior withdrawals.
//
//go:generate go tool mockgen -typed -destination=mocks/balance_service.go -package=mocks . BalanceService
type BalanceService interface {
	GetBalance(userID domain.UserID) (float64, error)
	GetTotalAmountWithdrawn(userID domain.UserID) (float64, error)
	GetWithdrawalTransactions(userID domain.UserID) ([]domain.WithdrawalTransaction, error)
	PayForOrderFromBalance(userID domain.UserID, orderID domain.OrderID, amount float64) error
}

// OrderService provides methods to upload new orders into the system
// and list prior user's orders.
//
//go:generate go tool mockgen -typed -destination=mocks/order_service.go -package=mocks . OrderService
type OrderService interface {
	CreateOrder(userID domain.UserID, orderID domain.OrderID) error
	GetOrders(userID domain.UserID) ([]domain.Order, error)
	GetAccruals(userID domain.UserID, orderIDs []domain.OrderID) (map[domain.OrderID]float64, error)
}
