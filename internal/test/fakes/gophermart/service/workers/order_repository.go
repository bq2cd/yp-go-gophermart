//nolint:revive,err113,wrapcheck,exhaustruct
package fakes

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/govalues/decimal"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

const (
	testOrderRepositoryBalanceDecimalScale = 16
)

type TestOrder struct {
	UserID  domain.UserID
	Status  domain.OrderStatus
	Accrual float64
}

type TestBalanceMap map[domain.UserID]float64
type TestOrderMap map[domain.OrderID]TestOrder

type TestOrderRepoData struct {
	Balances TestBalanceMap
	Orders   TestOrderMap
}

func NewTestOrderRepoData() TestOrderRepoData {
	return TestOrderRepoData{
		Balances: TestBalanceMap{},
		Orders:   TestOrderMap{},
	}
}

func (d *TestOrderRepoData) Merge(other TestOrderRepoData) {
	maps.Copy(d.Balances, other.Balances)
	maps.Copy(d.Orders, other.Orders)
}

func (d *TestOrderRepoData) ExpectEqual(other TestOrderRepoData) {
	ginkgo.GinkgoHelper()

	gomega.Expect(d.Balances).To(gomega.Equal(other.Balances))
	gomega.Expect(d.Orders).To(gomega.Equal(other.Orders))
}

/////////////////////////////////////////////////////////////////////////////////

type TestOrderRepoError struct {
	GetOrderStatus      TestError
	MarkOrderProcessing TestError
	MarkOrderProcessed  TestError
	MarkOrderInvalid    TestError
}

type TestOrderRepoErrorMap map[domain.OrderID]TestOrderRepoError

/////////////////////////////////////////////////////////////////////////////////

type TestOrderRepoDelay struct {
	GetOrderStatus      time.Duration
	MarkOrderProcessing time.Duration
	MarkOrderProcessed  time.Duration
	MarkOrderInvalid    time.Duration
}

type TestOrderRepoDelayMap map[domain.OrderID]TestOrderRepoDelay

/////////////////////////////////////////////////////////////////////////////////

var _ workers.OrderRepository = (*TestOrderRepository)(nil)

type TestOrderRepository struct {
	mu     sync.RWMutex
	level  slog.Level
	data   TestOrderRepoData
	errs   TestOrderRepoErrorMap
	delays TestOrderRepoDelayMap
}

func NewTestOrderRepository() *TestOrderRepository {
	return &TestOrderRepository{
		level:  testutil.LevelTrace,
		data:   TestOrderRepoData{},
		errs:   TestOrderRepoErrorMap{},
		delays: TestOrderRepoDelayMap{},
	}
}

func (r *TestOrderRepository) Setup(data TestOrderRepoData, errs TestOrderRepoErrorMap, delays TestOrderRepoDelayMap) {
	r.data = data
	r.errs = errs
	r.delays = delays
}

func (r *TestOrderRepository) GetData() *TestOrderRepoData {
	return &r.data
}

func (r *TestOrderRepository) GetOrderStatus(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) (domain.OrderStatus, error) {
	rtl := testutil.NewReturnLogger2[domain.OrderStatus, error](r.level, userID, orderID)

	if ctx.Err() != nil {
		return rtl.Log(domain.OrderStatusInvalid, ctx.Err())
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	order, err := r.getOrder(userID, orderID)
	if err != nil {
		return rtl.Log(domain.OrderStatusInvalid, err)
	}

	delay, ok := r.delays[orderID]
	if ok {
		time.Sleep(delay.GetOrderStatus)
	}

	err = r.errs[orderID].GetOrderStatus.Next()
	if err != nil {
		return rtl.Log(domain.OrderStatusInvalid, err)
	}

	return rtl.Log(order.Status, nil)
}

func (r *TestOrderRepository) MarkOrderInvalid(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) error {
	rtl := testutil.NewReturnLogger[error](r.level, userID, orderID)

	if ctx.Err() != nil {
		return rtl.Log(ctx.Err())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	order, err := r.getOrder(userID, orderID)
	if err != nil {
		return rtl.Log(err)
	}

	if order.Status == domain.OrderStatusProcessed {
		return rtl.Log(fmt.Errorf("cannot mark processed order (%v) as invalid", order))
	}

	delay, ok := r.delays[orderID]
	if ok {
		time.Sleep(delay.MarkOrderInvalid)
	}

	err = r.errs[orderID].MarkOrderInvalid.Next()
	if err != nil {
		return rtl.Log(err)
	}

	order.Status = domain.OrderStatusInvalid
	r.data.Orders[orderID] = order

	return rtl.Log(nil)
}

func (r *TestOrderRepository) MarkOrderProcessing(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
) error {
	rtl := testutil.NewReturnLogger[error](r.level, userID, orderID)

	if ctx.Err() != nil {
		return rtl.Log(ctx.Err())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	order, err := r.getOrder(userID, orderID)
	if err != nil {
		return rtl.Log(err)
	}

	//nolint:exhaustive
	switch order.Status {
	case domain.OrderStatusInvalid, domain.OrderStatusProcessed:
		return rtl.Log(
			fmt.Errorf("cannot mark invalid or processed order (%v) as processing", order),
		)
	case domain.OrderStatusProcessing:
		return nil
	}

	delay, ok := r.delays[orderID]
	if ok {
		time.Sleep(delay.MarkOrderProcessing)
	}

	err = r.errs[orderID].MarkOrderProcessing.Next()
	if err != nil {
		return rtl.Log(err)
	}

	order.Status = domain.OrderStatusProcessing
	r.data.Orders[orderID] = order

	return rtl.Log(nil)
}

func (r *TestOrderRepository) MarkOrderProcessed(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	amount float64,
) error {
	rtl := testutil.NewReturnLogger[error](r.level, userID, orderID, amount)

	if ctx.Err() != nil {
		return rtl.Log(ctx.Err())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	order, err := r.getOrder(userID, orderID)
	if err != nil {
		return rtl.Log(err)
	}

	//nolint:exhaustive
	switch order.Status {
	case domain.OrderStatusInvalid:
		return rtl.Log(fmt.Errorf("cannot mark invalid order (%v) as processed", order))
	case domain.OrderStatusProcessed:
		return nil
	}

	delay, ok := r.delays[orderID]
	if ok {
		time.Sleep(delay.MarkOrderProcessed)
	}

	err = r.errs[orderID].MarkOrderProcessed.Next()
	if err != nil {
		return rtl.Log(err)
	}

	order.Status = domain.OrderStatusProcessed
	order.Accrual = amount
	r.data.Orders[orderID] = order

	err = r.incrementBalance(userID, amount)
	if err != nil {
		return rtl.Log(err)
	}

	return rtl.Log(nil)
}

func (r *TestOrderRepository) getOrder(userID domain.UserID, orderID domain.OrderID) (TestOrder, error) {
	order, ok := r.data.Orders[orderID]
	if !ok {
		return TestOrder{}, domain.ErrOrderNotFound
	}

	err := r.validateUser(order, userID)
	if err != nil {
		return TestOrder{}, err
	}

	return order, nil
}

func (r *TestOrderRepository) validateUser(order TestOrder, expectedUserID domain.UserID) error {
	if order.UserID != expectedUserID {
		return fmt.Errorf("user ID mismatch: %v (order) != %v (expected)", order.UserID, expectedUserID)
	}

	return nil
}

func (r *TestOrderRepository) incrementBalance(userID domain.UserID, amount float64) error {
	increment, err := decimal.NewFromFloat64(amount)
	if err != nil {
		return fmt.Errorf("cannot convert amount to decimal: %w", err)
	}

	oldBalance, err := decimal.NewFromFloat64(r.data.Balances[userID])
	if err != nil {
		return fmt.Errorf("cannot convert old balance to decimal: %w", err)
	}

	newBalance, err := oldBalance.AddExact(increment, testOrderRepositoryBalanceDecimalScale)
	if err != nil {
		return fmt.Errorf("cannot add amount to balance exactly: %w", err)
	}

	result, _ := newBalance.Float64()

	r.data.Balances[userID] = result

	return nil
}
