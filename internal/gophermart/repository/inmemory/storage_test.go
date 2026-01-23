package inmemory_test

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/inmemory"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

// Ensure [inmemory.Storage] implements all necessary interfaces.
var (
	_ service.UserRepository    = (*inmemory.Storage)(nil)
	_ service.BalanceRepository = (*inmemory.Storage)(nil)
	_ service.OrderRepository   = (*inmemory.Storage)(nil)
	_ workers.OrderRepository   = (*inmemory.Storage)(nil)
)
