package inmemory

import (
	"sync"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// Storage implements multiple interfaces required by [service] packages.
// E.g. [service.UserRepository], [service.OrderRepository], [service.BalanceRepository].
// It also implements [workers.OrderRepository] required for [service.OrderProcessor].
type Storage struct {
	mu          sync.RWMutex
	users       map[domain.UserID]User
	orders      map[domain.OrderID]Order
	withdrawals map[domain.UserID][]domain.WithdrawalTransaction
}

// NewStorage creates an instance of [Storage].
func NewStorage() *Storage {
	return &Storage{
		mu:          sync.RWMutex{},
		users:       map[domain.UserID]User{},
		orders:      map[domain.OrderID]Order{},
		withdrawals: map[domain.UserID][]domain.WithdrawalTransaction{},
	}
}
