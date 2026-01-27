//nolint:revive,wrapcheck,exhaustruct
package fakes

import (
	"context"
	"log/slog"
	"maps"
	"sync"
	"time"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

type TestAccrualOrder struct {
	Status  accdomain.OrderStatus
	Accrual float64
}

type TestAccrualData map[domain.OrderID]TestAccrualOrder

func (m TestAccrualData) Merge(other TestAccrualData) {
	maps.Copy(m, other)
}

type TestAccrualError struct {
	GetOrderStatus TestError
}

type TestAccrualErrorMap map[domain.OrderID]TestAccrualError

func (m TestAccrualErrorMap) Merge(other TestAccrualErrorMap) {
	maps.Copy(m, other)
}

type TestAccrualDelay struct {
	GetOrderStatus time.Duration
}

type TestAccrualDelayMap map[domain.OrderID]TestAccrualDelay

/////////////////////////////////////////////////////////////////////////////////

var _ workers.AccrualClient = (*TestAccrualClient)(nil)

type TestAccrualClient struct {
	mu     sync.RWMutex
	level  slog.Level
	data   TestAccrualData
	errs   TestAccrualErrorMap
	delays TestAccrualDelayMap
}

func NewTestAccrualClient() *TestAccrualClient {
	return &TestAccrualClient{
		level:  testutil.LevelTrace,
		data:   TestAccrualData{},
		errs:   TestAccrualErrorMap{},
		delays: TestAccrualDelayMap{},
	}
}

func (c *TestAccrualClient) Setup(data TestAccrualData, errs TestAccrualErrorMap, delays TestAccrualDelayMap) {
	c.data = data
	c.errs = errs
	c.delays = delays
}

func (c *TestAccrualClient) GetOrderStatus(ctx context.Context, orderID accdomain.OrderID) (accdomain.Order, error) {
	rtl := testutil.NewReturnLogger2[accdomain.Order, error](c.level, orderID)

	if ctx.Err() != nil {
		return rtl.Log(accdomain.Order{}, ctx.Err())
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	order, hasOrder := c.data[domain.OrderID(orderID)]
	if !hasOrder {
		return rtl.Log(accdomain.Order{}, accdomain.ErrOrderNotFound)
	}

	delay, ok := c.delays[domain.OrderID(orderID)]
	if ok {
		time.Sleep(delay.GetOrderStatus)
	}

	err := c.errs[domain.OrderID(orderID)].GetOrderStatus.Next()
	if err != nil {
		return rtl.Log(accdomain.Order{}, err)
	}

	accOrder := accdomain.Order{
		ID:            orderID,
		Status:        order.Status,
		AccrualPoints: order.Accrual,
	}

	return rtl.Log(accOrder, nil)
}
