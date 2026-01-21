package integration

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

type Storage interface {
	service.UserRepository
	service.BalanceRepository
	service.OrderRepository
	workers.OrderRepository
}
