package workers

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

const (
	orderProcessorDefaultPerEventTimeout = 2 * time.Second
	orderProcessorDefaultShutdownTimeout = 5 * time.Second
)

// OrderProcessor implements [service.OrderProcessor].
type OrderProcessor struct {
	orderRepo       OrderRepository
	queue           OrderQueue
	accrualClient   AccrualClient
	perEventTimeout time.Duration
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
	options ...option.Option[OrderProcessor],
) *OrderProcessor {
	processor := &OrderProcessor{
		orderRepo:       orderRepository,
		queue:           orderQueue,
		accrualClient:   accrualClient,
		perEventTimeout: orderProcessorDefaultPerEventTimeout,
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

// WithOrderProcessorPerEventTimeout returns an option to configure
// processing timeout per order event.
func WithOrderProcessorPerEventTimeout(timeout time.Duration) option.Option[OrderProcessor] {
	return func(p *OrderProcessor) {
		p.perEventTimeout = timeout
	}
}

// WithOrderProcessorShutdownTimeout returns an option to configure
// total timeout for a graceful shutdown of the [OrderProcessor].
func WithOrderProcessorShutdownTimeout(timeout time.Duration) option.Option[OrderProcessor] {
	return func(p *OrderProcessor) {
		p.shutdownTimeout = timeout
	}
}

// EnqueueOrder puts given order ID into in-memory queue for further processing.
func (p *OrderProcessor) EnqueueOrder(userID domain.UserID, orderID domain.OrderID) bool {
	event := OrderEvent{
		userID:  userID,
		orderID: orderID,
	}

	return p.enqueueEvent(event)
}

// Run launches an infinite loop which consumes [OrderEvent] events from
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
// processing of all enqueued events after it has been shutdown.
// It is primarily used in tests.
func (p *OrderProcessor) HasFinished() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.queue.Len() == 0
}

func (p *OrderProcessor) enqueueEvent(event OrderEvent) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return false
	}

	p.queue.PushBack(event)
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
		event OrderEvent
		ok    bool
	)

	for {
		event, ok = p.nextEvent()
		if !ok {
			break
		}

		p.processEvent(ctx, event)
	}
}

func (p *OrderProcessor) nextEvent() (OrderEvent, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.queue.Len() == 0 {
		//nolint:exhaustruct
		return OrderEvent{}, false
	}

	return p.queue.PopFront(), true
}

// shutdown processes any remaining events in the [OrderQueue] while respecting
// [shutdownTimeout].
func (p *OrderProcessor) shutdown(ctx context.Context) {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), p.shutdownTimeout)
	defer cancel()

	p.processQueue(shutdownCtx)
}

func (p *OrderProcessor) processEvent(baseCtx context.Context, event OrderEvent) {
	ctx, cancel := context.WithTimeout(baseCtx, p.perEventTimeout)
	defer cancel()

	if !p.shouldProcessEvent(ctx, event) {
		return
	}

	accOrder, err := p.accrualClient.GetOrderStatus(ctx, accdomain.OrderID(event.orderID))
	if err != nil {
		p.processAccrualClientError(ctx, event, err)

		return
	}

	p.processAccrualClientOrder(ctx, event.userID, accOrder)
}

func (p *OrderProcessor) shouldProcessEvent(ctx context.Context, event OrderEvent) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	status, err := p.orderRepo.GetOrderStatus(ctx, event.userID, event.orderID)
	if err != nil {
		p.processGetOrderStatusError(event, err)

		return false
	}

	return p.shouldProcessOrderStatus(ctx, event, status)
}

func (p *OrderProcessor) processGetOrderStatusError(event OrderEvent, err error) {
	if errors.Is(err, domain.ErrOrderNotFound) {
		return
	}

	p.enqueueEvent(event)
}

func (p *OrderProcessor) shouldProcessOrderStatus(
	ctx context.Context,
	event OrderEvent,
	status domain.OrderStatus,
) bool {
	var isEligible bool

	switch status {
	case domain.OrderStatusNew, domain.OrderStatusProcessing:
		isEligible = true
	case domain.OrderStatusInvalid, domain.OrderStatusProcessed:
		isEligible = false
	}

	slog.DebugContext(ctx, "order eligibility for processing",
		slog.Group("order",
			slog.Uint64("id", uint64(event.orderID)),
			slog.String("user", string(event.userID)),
			slog.Int("status", int(status)),
		),
		slog.Bool("is_eligible", isEligible),
	)

	return isEligible
}

func (p *OrderProcessor) processAccrualClientError(_ context.Context, event OrderEvent, _ error) {
	p.retryEvent(event)
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
	event := OrderEvent{
		userID:  userID,
		orderID: orderID,
	}

	p.retryEvent(event)
}

func (p *OrderProcessor) retryEvent(event OrderEvent) {
	p.enqueueEvent(event)
}
