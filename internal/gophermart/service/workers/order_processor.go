package workers

import (
	"context"
	"errors"
	"sync"
	"time"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

const (
	orderProcessorDefaultPerItemTimeout  = 2 * time.Second
	orderProcessorDefaultShutdownTimeout = 5 * time.Second
)

// OrderProcessorOption defines an option type
// that can be used to configure [OrderProcessor].
type OrderProcessorOption func(*OrderProcessor)

// OrderProcessor implements [service.OrderProcessor].
type OrderProcessor struct {
	orderRepo       OrderRepository
	queue           OrderQueue
	accrualClient   AccrualClient
	perItemTimeout  time.Duration
	shutdownTimeout time.Duration
	mu              sync.Mutex
	notifyCh        chan struct{}
	isClosed        bool
}

// NewOrderProcessor creates an instance of [OrderProcessor].
func NewOrderProcessor(
	orderRepository OrderRepository,
	orderQueue OrderQueue,
	accrualClient AccrualClient,
	options ...OrderProcessorOption,
) *OrderProcessor {
	processor := &OrderProcessor{
		orderRepo:       orderRepository,
		queue:           orderQueue,
		accrualClient:   accrualClient,
		perItemTimeout:  orderProcessorDefaultPerItemTimeout,
		shutdownTimeout: orderProcessorDefaultShutdownTimeout,
		mu:              sync.Mutex{},
		notifyCh:        make(chan struct{}, 1),
		isClosed:        false,
	}
	for _, opt := range options {
		opt(processor)
	}

	return processor
}

// WithOrderProcessorPerItemTimeout returns an option to configure
// processing timeout per item.
func WithOrderProcessorPerItemTimeout(timeout time.Duration) OrderProcessorOption {
	return func(p *OrderProcessor) {
		p.perItemTimeout = timeout
	}
}

// WithOrderProcessorShutdownTimeout returns an option to configure
// total timeout for a graceful shutdown of the [OrderProcessor].
func WithOrderProcessorShutdownTimeout(timeout time.Duration) OrderProcessorOption {
	return func(p *OrderProcessor) {
		p.shutdownTimeout = timeout
	}
}

// EnqueueOrder puts given order ID into in-memory queue for further processing.
func (p *OrderProcessor) EnqueueOrder(userID domain.UserID, orderID domain.OrderID) bool {
	item := OrderItem{
		UserID:  userID,
		OrderID: orderID,
	}

	return p.enqueueItem(item)
}

// Run launches an infinite loop which consumes [OrderItem] items from
// an internal channel and processes them.
// This method should be called from a goroutine.
func (p *OrderProcessor) Run(ctx context.Context) {
	p.mainLoop(ctx)
	p.shutdown(ctx)
}

// IsClosed returns [true] when [OrderProcessor] has started
// a shutdown process and has stopped accepting new orders.
// It is primarily used in tests.
func (p *OrderProcessor) IsClosed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.isClosed
}

// HasFinished returns [true] when [OrderProcessor] finishes
// processing of all enqueued items after it has been shutdown.
// It is primarily used in tests.
func (p *OrderProcessor) HasFinished() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.queue.Len() == 0
}

func (p *OrderProcessor) enqueueItem(item OrderItem) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return false
	}

	p.queue.PushBack(item)
	p.notifyMainLoop()

	return true
}

func (p *OrderProcessor) notifyMainLoop() {
	// Ensure we never block on sending to notify channel.
	select {
	case <-p.notifyCh:
	default:
	}

	p.notifyCh <- struct{}{}
}

func (p *OrderProcessor) mainLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			p.setIsClosed()

			break loop
		case <-p.notifyCh:
			p.processQueue(ctx)
		}
	}
}

func (p *OrderProcessor) setIsClosed() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isClosed = true
}

func (p *OrderProcessor) processQueue(ctx context.Context) {
	var (
		item OrderItem
		ok   bool
	)

	for {
		item, ok = p.nextItem()
		if !ok {
			break
		}

		p.processItem(ctx, item)
	}
}

func (p *OrderProcessor) nextItem() (OrderItem, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.queue.Len() == 0 {
		//nolint:exhaustruct
		return OrderItem{}, false
	}

	return p.queue.PopFront(), true
}

// shutdown processes any remaining items in the [OrderQueue] while respecting
// [shutdownTimeout].
func (p *OrderProcessor) shutdown(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), p.shutdownTimeout)
	defer cancel()

	p.processQueue(shutdownCtx)
}

func (p *OrderProcessor) processItem(baseCtx context.Context, item OrderItem) {
	ctx, cancel := context.WithTimeout(baseCtx, p.perItemTimeout)
	defer cancel()

	if !p.shouldProcessItem(ctx, item) {
		return
	}

	accOrder, err := p.accrualClient.GetOrderStatus(ctx, accdomain.OrderID(item.OrderID))
	if err != nil {
		p.processAccrualClientError(ctx, item, err)

		return
	}

	p.processAccrualClientOrder(ctx, item.UserID, accOrder)
}

func (p *OrderProcessor) shouldProcessItem(ctx context.Context, item OrderItem) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	status, err := p.orderRepo.GetOrderStatus(ctx, item.UserID, item.OrderID)
	if err != nil {
		p.processGetOrderStatusError(item, err)

		return false
	}

	return p.shouldProcessOrderStatus(ctx, item, status)
}

func (p *OrderProcessor) processGetOrderStatusError(item OrderItem, err error) {
	if errors.Is(err, domain.ErrOrderNotFound) {
		return
	}

	p.enqueueItem(item)
}

func (p *OrderProcessor) shouldProcessOrderStatus(ctx context.Context, item OrderItem, status domain.OrderStatus) bool {
	switch status {
	case domain.OrderStatusNew:
		return p.markOrderProcessing(ctx, item.UserID, item.OrderID)
	case domain.OrderStatusProcessing:
		return true
	case domain.OrderStatusInvalid, domain.OrderStatusProcessed:
		return false
	default:
		return false
	}
}

func (p *OrderProcessor) processAccrualClientError(ctx context.Context, item OrderItem, err error) {
	if errors.Is(err, accdomain.ErrOrderNotFound) {
		p.markOrderInvalid(ctx, item.UserID, item.OrderID)

		return
	}

	p.retryItem(item)
}

func (p *OrderProcessor) processAccrualClientOrder(
	ctx context.Context,
	userID domain.UserID,
	accOrder accdomain.Order,
) {
	orderID := domain.OrderID(accOrder.ID)

	switch accOrder.Status {
	case accdomain.OrderStatusRegistered, accdomain.OrderStatusProcessing:
		p.markOrderProcessing(ctx, userID, orderID)
	case accdomain.OrderStatusProcessed:
		p.markOrderProcessed(ctx, userID, orderID, accOrder.AccrualPoints)
	case accdomain.OrderStatusInvalid:
		p.markOrderInvalid(ctx, userID, orderID)
	default:
		p.markOrderInvalid(ctx, userID, orderID)
	}
}

func (p *OrderProcessor) markOrderProcessing(ctx context.Context, userID domain.UserID, orderID domain.OrderID) bool {
	err := p.orderRepo.MarkOrderProcessing(ctx, userID, orderID)
	if err != nil {
		p.retryOrder(userID, orderID)

		return false
	}

	return true
}

func (p *OrderProcessor) markOrderInvalid(ctx context.Context, userID domain.UserID, orderID domain.OrderID) {
	err := p.orderRepo.MarkOrderInvalid(ctx, userID, orderID)
	if err != nil {
		p.retryOrder(userID, orderID)
	}
}

func (p *OrderProcessor) markOrderProcessed(
	ctx context.Context,
	userID domain.UserID,
	orderID domain.OrderID,
	accrual float64,
) {
	err := p.orderRepo.MarkOrderProcessed(ctx, userID, orderID, accrual)
	if err != nil {
		p.retryOrder(userID, orderID)
	}
}

func (p *OrderProcessor) retryOrder(userID domain.UserID, orderID domain.OrderID) {
	item := OrderItem{
		UserID:  userID,
		OrderID: orderID,
	}

	p.retryItem(item)
}

func (p *OrderProcessor) retryItem(item OrderItem) {
	p.enqueueItem(item)
}
