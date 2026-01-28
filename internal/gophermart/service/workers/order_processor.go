package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

const (
	orderProcessorDefaultPerEventTimeout = 2 * time.Second
	orderProcessorDefaultShutdownTimeout = 5 * time.Second
	orderProcessorDefaultWorkerPoolSize  = 2
)

// OrderProcessor implements [service.OrderProcessor].
type OrderProcessor struct {
	wg                     sync.WaitGroup
	runCh                  chan struct{}
	callbackCh             <-chan orderEventResult
	config                 *orderProcessorConfig
	state                  *orderProcessorState
	workerPool             *orderEventWorkerPool
	getProcessableOrdersFn func(context.Context) (map[domain.UserID][]domain.OrderID, error)
}

// NewOrderProcessor creates an instance of [OrderProcessor].
func NewOrderProcessor(
	orderRepository OrderRepository,
	orderQueue OrderQueue,
	accrualClient AccrualClient,
	options ...option.Option[OrderProcessor],
) *OrderProcessor {
	config := &orderProcessorConfig{
		ShutdownTimeout:              orderProcessorDefaultShutdownTimeout,
		DelayConfig:                  NewOrderEventDelayConfig(),
		EnableOrderPreloadingOnStart: true,
	}
	state := newOrderProcessorState(orderQueue)
	workerPool := newOrderEventWorkerPool(orderRepository, accrualClient)

	processor := &OrderProcessor{
		wg:                     sync.WaitGroup{},
		runCh:                  make(chan struct{}, 1),
		callbackCh:             nil,
		config:                 config,
		state:                  state,
		workerPool:             workerPool,
		getProcessableOrdersFn: orderRepository.GetProcessableOrdersPerUser,
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
		p.workerPool.SetPerEventTimeout(timeout)
	}
}

// WithOrderProcessorShutdownTimeout returns an option to configure
// total timeout for a graceful shutdown of the [OrderProcessor].
func WithOrderProcessorShutdownTimeout(timeout time.Duration) option.Option[OrderProcessor] {
	return func(p *OrderProcessor) {
		p.config.ShutdownTimeout = timeout
	}
}

// WithOrderProcessorWorkerPoolSize returns an option to configure
// worker pool size for event processing.
func WithOrderProcessorWorkerPoolSize(size uint) option.Option[OrderProcessor] {
	return func(p *OrderProcessor) {
		p.workerPool.SetPoolSize(size)
	}
}

// WithOrderProcessorDelayConfig returns an option to configure delays for retryable events.
// The config includes minimum (base) delay, maximum possible delay,
// and max jitter.
// These values will be the same for all types of events, although
// concrete retry strategy (fixed, backoff, random) will be decided
// internally based on event processing results.
func WithOrderProcessorDelayConfig(config retry.DelayContext) option.Option[OrderProcessor] {
	return func(p *OrderProcessor) {
		p.config.DelayConfig = config
	}
}

// WithOrderProcessorEnableOrderPreloadingOnStart returns an option to configure default behavior
// of [OrderProcessor] when [Start] method is invoked.
// By default, it will try to load all processable orders from [OrderRepository] and enqueue them
// for processing.
// This behavior is useful on application startup.
// For testing, however, one might want to disable automatic preloading and enqueue orders
// manually for better flow control.
func WithOrderProcessorEnableOrderPreloadingOnStart(enable bool) option.Option[OrderProcessor] {
	return func(p *OrderProcessor) {
		p.config.EnableOrderPreloadingOnStart = enable
	}
}

// EnqueueOrder puts given order ID into in-memory queue for further processing.
func (p *OrderProcessor) EnqueueOrder(userID domain.UserID, orderID domain.OrderID) bool {
	event := p.newEvent(userID, orderID)

	return p.state.EnqueueEvent(event)
}

// Run launches an infinite loop which consumes [OrderEvent] events from
// an internal channel and processes them.
// This method should be called from a goroutine.
func (p *OrderProcessor) Run(ctx context.Context) {
	p.runCh <- struct{}{}

	p.state.SetIsClosed(false)

	p.startEnqueueingProcessableOrders(ctx)
	p.startWorkers(ctx)

	p.mainLoop(ctx)

	p.state.SetIsClosed(true)

	p.shutdown(ctx)

	<-p.runCh
}

// IsClosed returns [true] when [OrderProcessor] has started
// a shutdown process and has stopped accepting new orders.
// It is primarily used in tests.
func (p *OrderProcessor) IsClosed() bool {
	return p.state.IsClosed()
}

// HasFinished returns [true] when [OrderProcessor] finishes
// processing of all enqueued events after it has been shutdown.
// It is primarily used in tests.
func (p *OrderProcessor) HasFinished() bool {
	return p.state.HasFinished()
}

// EnableOrderPreloadingOnStart allows to change pre-configured behavior
// of [OrderProcessor] when [Run] method is invoked.
// This method will only have an effect if called before [Run] is called.
// It is primarily useful for testing to control multiple invocations
// of [OrderProcessor].
func (p *OrderProcessor) EnableOrderPreloadingOnStart(enable bool) {
	p.config.EnableOrderPreloadingOnStart = enable
}

func (p *OrderProcessor) newEvent(userID domain.UserID, orderID domain.OrderID) OrderEvent {
	return OrderEvent{
		userID:       userID,
		orderID:      orderID,
		processAfter: time.Now(),
		retries:      0,
		delayConfig:  p.config.DelayConfig,
	}
}

func (p *OrderProcessor) startEnqueueingProcessableOrders(ctx context.Context) {
	if !p.config.EnableOrderPreloadingOnStart {
		return
	}

	p.wg.Add(1)

	go p.enqueueProcessableOrders(ctx)
}

func (p *OrderProcessor) enqueueProcessableOrders(ctx context.Context) {
	defer p.wg.Done()

	slog.DebugContext(ctx, "processor: preloading processable orders")

	ordersPerUser, err := p.getProcessableOrdersFn(ctx)
	if err != nil {
		return
	}

	for userID, orders := range ordersPerUser {
		for _, orderID := range orders {
			p.EnqueueOrder(userID, orderID)
		}
	}
}

func (p *OrderProcessor) startWorkers(ctx context.Context) {
	callbackCh := make(chan orderEventResult)

	p.callbackCh = callbackCh

	p.startCallbackProcessing(ctx)

	p.workerPool.Start(ctx, callbackCh)
}

func (p *OrderProcessor) startCallbackProcessing(ctx context.Context) {
	p.wg.Add(1)

	go p.runCallbackProcessing(ctx)
}

func (p *OrderProcessor) runCallbackProcessing(ctx context.Context) {
	defer p.wg.Done()

	for result := range p.callbackCh {
		p.processCallback(ctx, result)
	}
}

func (p *OrderProcessor) processCallback(ctx context.Context, result orderEventResult) {
	slog.DebugContext(ctx, "processor: callback",
		slog.Any("event", result.OrderEvent),
		slog.Any("error", result.err),
	)

	p.processEventResult(ctx, result)
}

func (p *OrderProcessor) processEventResult(ctx context.Context, result orderEventResult) {
	defer p.state.DecrementEventsInFlight()

	if hasContextExpired(ctx) {
		return
	}

	event, retryable := result.GetRetryableEvent()
	if !retryable {
		return
	}

	p.state.EnqueueEvent(event)
}

func (p *OrderProcessor) mainLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-p.state.NotifyC():
			p.processQueue(ctx)
		}
	}

	slog.DebugContext(ctx, "processor: main loop completed")
}

func (p *OrderProcessor) processQueue(ctx context.Context) {
	for !hasContextExpired(ctx) {
		event, ok := p.state.NextEvent()
		if !ok {
			break
		}

		p.processEvent(ctx, event)
	}
}

func (p *OrderProcessor) processEvent(ctx context.Context, event OrderEvent) {
	if time.Now().After(event.processAfter) {
		p.sendEventToWorkerPool(ctx, event)
	} else {
		p.sendEventBackToQueue(ctx, event)
	}
}

func (p *OrderProcessor) sendEventToWorkerPool(ctx context.Context, event OrderEvent) {
	slog.DebugContext(ctx, "processor: sending to worker pool",
		slog.Any("event", event),
		slog.Any("ctx", ctx),
	)

	p.workerPool.TakeEvent(event)
}

func (p *OrderProcessor) sendEventBackToQueue(ctx context.Context, event OrderEvent) {
	p.wg.Add(1)

	go p.waitForEventToBecomeProcessable(ctx, event)
}

func (p *OrderProcessor) waitForEventToBecomeProcessable(ctx context.Context, event OrderEvent) {
	defer p.wg.Done()

	defer p.state.DecrementEventsInFlight()

	timer := time.NewTimer(time.Until(event.processAfter))

	slog.DebugContext(ctx, "processor: event waiter started",
		slog.Any("event", event),
		slog.Any("ctx", ctx),
	)

	select {
	case <-ctx.Done():
		timer.Stop()
	case <-timer.C:
		if !hasContextExpired(ctx) {
			p.state.EnqueueEvent(event)
		}
	}

	slog.DebugContext(ctx, "processor: event waiter finished",
		slog.Any("event", event),
		slog.Any("ctx", ctx),
	)
}

// shutdown processes any remaining events in the [OrderQueue] while respecting [shutdownTimeout].
func (p *OrderProcessor) shutdown(baseCtx context.Context) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(baseCtx), p.config.ShutdownTimeout)
	defer cancel()

	slog.DebugContext(ctx, "processor: shutdown started",
		slog.Int("queue_size", p.state.QueueSize()),
		slog.Any("ctx", ctx),
	)

	p.waitForWorkersToFinish()

	p.startWorkers(ctx)
	p.drainQueue(ctx)

	slog.DebugContext(ctx, "processor: shutdown queue drained")

	p.waitForWorkersToFinish()

	slog.DebugContext(ctx, "processor: shutdown finished")
}

func (p *OrderProcessor) waitForWorkersToFinish() {
	p.workerPool.Wait()
	p.wg.Wait()
}

func (p *OrderProcessor) drainQueue(ctx context.Context) {
	for {
		event, ok := p.state.NextEvent()
		if !ok {
			break
		}

		if hasContextExpired(ctx) {
			p.processCallback(ctx, orderEventResult{
				OrderEvent: event,
				err:        ErrOrderEventDiscarded,
			})

			continue
		}

		p.processEvent(ctx, event)
	}
}
