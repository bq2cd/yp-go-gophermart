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
	mu              sync.Mutex
	wg              sync.WaitGroup
	queue           OrderQueue
	notifyCh        chan struct{}
	callbackCh      <-chan orderEventResult
	isClosed        bool
	workerPool      *orderEventWorkerPool
	shutdownTimeout time.Duration
	eventsInFlight  uint64
	delayConfig     retry.DelayContext
}

// NewOrderProcessor creates an instance of [OrderProcessor].
func NewOrderProcessor(
	orderRepository OrderRepository,
	orderQueue OrderQueue,
	accrualClient AccrualClient,
	options ...option.Option[OrderProcessor],
) *OrderProcessor {
	workerPool := newOrderEventWorkerPool(orderRepository, accrualClient)

	processor := &OrderProcessor{
		mu:              sync.Mutex{},
		wg:              sync.WaitGroup{},
		queue:           orderQueue,
		notifyCh:        make(chan struct{}, 1),
		callbackCh:      nil,
		isClosed:        false,
		workerPool:      workerPool,
		shutdownTimeout: orderProcessorDefaultShutdownTimeout,
		eventsInFlight:  0,
		delayConfig:     NewOrderEventDelayConfig(),
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
		p.shutdownTimeout = timeout
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
		p.delayConfig = config
	}
}

// EnqueueOrder puts given order ID into in-memory queue for further processing.
func (p *OrderProcessor) EnqueueOrder(userID domain.UserID, orderID domain.OrderID) bool {
	event := p.newEvent(userID, orderID)

	return p.enqueueEvent(event)
}

// Run launches an infinite loop which consumes [OrderEvent] events from
// an internal channel and processes them.
// This method should be called from a goroutine.
func (p *OrderProcessor) Run(ctx context.Context) {
	p.startWorkers(ctx)
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

	isEmptyQueue := p.queue.Len() == 0
	hasProcessingFinished := p.eventsInFlight == 0

	return isEmptyQueue && hasProcessingFinished
}

func (p *OrderProcessor) newEvent(userID domain.UserID, orderID domain.OrderID) OrderEvent {
	return OrderEvent{
		userID:       userID,
		orderID:      orderID,
		processAfter: time.Now(),
		retries:      0,
		delayConfig:  p.delayConfig,
	}
}

func (p *OrderProcessor) enqueueEvent(event OrderEvent) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isClosed {
		return false
	}

	slog.Debug("processor: enqueue",
		slog.Any("event", event),
		slog.Int("queue_size", p.queue.Len()),
	)

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
	defer p.decrementEventsInFlight()

	if hasContextExpired(ctx) {
		return
	}

	event, retryable := result.GetRetryableEvent()
	if !retryable {
		return
	}

	p.enqueueEvent(event)
}

func (p *OrderProcessor) decrementEventsInFlight() {
	p.mu.Lock()
	defer p.mu.Unlock()

	inflight := p.eventsInFlight

	if p.eventsInFlight > 0 {
		p.eventsInFlight--
	}

	slog.Debug("processor: inflight--",
		slog.Uint64("before", inflight),
		slog.Uint64("after", p.eventsInFlight),
		slog.Int("queue_size", p.queue.Len()),
	)
}

func (p *OrderProcessor) mainLoop(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-p.notifyCh:
			p.processQueue(ctx)
		}
	}

	p.setIsClosed()

	slog.DebugContext(ctx, "processor: main loop completed")
}

func (p *OrderProcessor) setIsClosed() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isClosed = true
}

func (p *OrderProcessor) processQueue(ctx context.Context) {
	for !hasContextExpired(ctx) {
		event, ok := p.nextEvent()
		if !ok {
			break
		}

		p.processEvent(ctx, event)
	}
}

func (p *OrderProcessor) nextEvent() (OrderEvent, bool) {
	var event OrderEvent

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.queue.Len() == 0 {
		return event, false
	}

	event = p.queue.PopFront()

	inflight := p.eventsInFlight
	p.eventsInFlight++

	slog.Debug("processor: inflight++",
		slog.Uint64("before", inflight),
		slog.Uint64("after", p.eventsInFlight),
		slog.Int("queue_size", p.queue.Len()),
	)

	return event, true
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

	defer p.decrementEventsInFlight()

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
			p.enqueueEvent(event)
		}
	}

	slog.DebugContext(ctx, "processor: event waiter finished",
		slog.Any("event", event),
		slog.Any("ctx", ctx),
	)
}

// shutdown processes any remaining events in the [OrderQueue] while respecting [shutdownTimeout].
func (p *OrderProcessor) shutdown(baseCtx context.Context) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(baseCtx), p.shutdownTimeout)
	defer cancel()

	slog.DebugContext(ctx, "processor: shutdown started",
		slog.Int("queue_size", p.queueSize()),
		slog.Any("ctx", ctx),
	)

	p.waitForWorkersToFinish(ctx)

	p.startWorkers(ctx)
	p.drainQueue(ctx)

	slog.DebugContext(ctx, "processor: shutdown queue drained")

	p.waitForWorkersToFinish(ctx)

	slog.DebugContext(ctx, "processor: shutdown finished")
}

func (p *OrderProcessor) queueSize() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.queue.Len()
}

func (p *OrderProcessor) waitForWorkersToFinish(ctx context.Context) {
	p.workerPool.Wait(ctx)
	p.wg.Wait()
}

func (p *OrderProcessor) drainQueue(ctx context.Context) {
	for {
		event, ok := p.nextEvent()
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
