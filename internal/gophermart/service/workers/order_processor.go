package workers

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

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
	wg         sync.WaitGroup
	runCh      chan struct{}
	callbackCh <-chan orderEventResult
	numWakeups atomic.Uint64
	config     *orderProcessorConfig
	state      *orderProcessorState
	workerPool *orderEventWorkerPool
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
		wg:         sync.WaitGroup{},
		runCh:      make(chan struct{}, 1),
		callbackCh: nil,
		numWakeups: atomic.Uint64{},
		config:     config,
		state:      state,
		workerPool: workerPool,
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
func WithOrderProcessorDelayConfig(config OrderEventDelayConfig) option.Option[OrderProcessor] {
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

// NewOrderArrived notifies main processing loop about new order.
func (p *OrderProcessor) NewOrderArrived(_ domain.UserID, _ domain.OrderID) {
	p.state.TriggerQueueProcessing()
}

// Run launches an infinite loop which consumes [OrderEvent] events from
// an internal channel and processes them.
// This method should be called from a goroutine.
func (p *OrderProcessor) Run(ctx context.Context) {
	p.runCh <- struct{}{}

	p.startEnqueueingProcessableOrders(ctx)
	p.startWorkers(ctx)

	p.mainLoop(ctx)

	p.shutdown(ctx)

	<-p.runCh
}

// HasFinished returns [true] when [OrderProcessor] finishes
// processing of all enqueued events after it has been shutdown.
// It is primarily used in tests.
func (p *OrderProcessor) HasFinished() bool {
	hasStarted := p.state.HasStarted()
	hasFinished := p.state.HasFinished()

	slog.Debug("processor: has finished yet?",
		slog.Bool("has_started", hasStarted),
		slog.Bool("no_events_inflight", hasFinished),
	)

	return hasStarted && hasFinished
}

// EnableOrderPreloadingOnStart allows to change pre-configured behavior
// of [OrderProcessor] when [Run] method is invoked.
// This method will only have an effect if called before [Run] is called.
// It is primarily useful for testing to control multiple invocations
// of [OrderProcessor].
func (p *OrderProcessor) EnableOrderPreloadingOnStart(enable bool) {
	p.config.EnableOrderPreloadingOnStart = enable
}

// NumWakeups returns a number of times when [OrderProcessor] got woken up to
// process the queue. This can happen when it got notified about a new order
// or when a pending order approached its processing time.
// This method is primarily exposed for testing purposes.
func (p *OrderProcessor) NumWakeups() uint64 {
	return p.numWakeups.Load()
}

func (p *OrderProcessor) startEnqueueingProcessableOrders(_ context.Context) {
	if !p.config.EnableOrderPreloadingOnStart {
		return
	}

	p.state.TriggerQueueProcessing()
}

func (p *OrderProcessor) startWorkers(ctx context.Context) {
	callbackCh := make(chan orderEventResult, 1)

	p.callbackCh = callbackCh

	p.startCallbackProcessing(ctx)

	p.workerPool.Start(ctx, callbackCh)
}

func (p *OrderProcessor) startCallbackProcessing(ctx context.Context) {
	p.wg.Add(1)

	go p.runCallbackProcessing(ctx)
}

func (p *OrderProcessor) mainLoop(ctx context.Context) {
	slog.DebugContext(ctx, "processor: main loop started")

	p.numWakeups.Store(0)
	p.state.Start()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-p.state.NotifyC():
			slog.DebugContext(ctx, "processor: queue processing triggered by notification")
			p.processQueue(ctx)
		case <-p.state.WakeupC():
			slog.DebugContext(ctx, "processor: queue processing triggered by wakeup timer")
			p.processQueue(ctx)
		}
	}

	slog.DebugContext(ctx, "processor: main loop finished")
}

func (p *OrderProcessor) processQueue(ctx context.Context) {
	p.numWakeups.Add(1)

	for !hasContextExpired(ctx) {
		event, ok := p.state.NextEvent(ctx, p.config.DelayConfig)
		if !ok {
			break
		}

		p.processEvent(ctx, event)
	}
}

func (p *OrderProcessor) processEvent(ctx context.Context, event OrderEvent) {
	p.sendEventToWorkerPool(ctx, event)
}

func (p *OrderProcessor) sendEventToWorkerPool(ctx context.Context, event OrderEvent) {
	slog.DebugContext(ctx, "processor: sending to worker pool",
		slog.Any("event", event),
		slog.Any("ctx", ctx),
	)

	p.workerPool.TakeEvent(event)
}

func (p *OrderProcessor) runCallbackProcessing(ctx context.Context) {
	defer p.wg.Done()

	slog.DebugContext(ctx, "processor: callback processing started")

	for result := range p.callbackCh {
		p.processCallback(ctx, result)
	}

	slog.DebugContext(ctx, "processor: callback processing finished")
}

func (p *OrderProcessor) processCallback(ctx context.Context, result orderEventResult) {
	slog.DebugContext(ctx, "processor: callback",
		slog.Any("event", result.OrderEvent),
		slog.Any("error", result.err),
	)

	defer p.state.DecrementEventsInFlight(result.OrderID)

	if hasContextExpired(ctx) {
		return
	}

	event, retryable := result.GetRetryableEvent()
	if !retryable {
		return
	}

	p.state.EnqueueEvent(ctx, event)
}

// shutdown processes any remaining events in the [OrderQueue] while respecting [shutdownTimeout].
func (p *OrderProcessor) shutdown(baseCtx context.Context) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(baseCtx), p.config.ShutdownTimeout)
	defer cancel()

	slog.DebugContext(ctx, "processor: shutdown started",
		slog.Any("ctx", ctx),
	)

	p.state.Stop()

	p.waitForWorkersToFinish()

	slog.DebugContext(ctx, "processor: shutdown finished")
}

func (p *OrderProcessor) waitForWorkersToFinish() {
	p.workerPool.Wait()
	p.wg.Wait()
}
