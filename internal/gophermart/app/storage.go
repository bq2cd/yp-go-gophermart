package app

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

// Storage represents a storage backend that provides all necessary
// repositories for the app.
type Storage interface {
	service.UserRepository
	service.BalanceRepository
	service.OrderRepository
	workers.OrderRepository
}
