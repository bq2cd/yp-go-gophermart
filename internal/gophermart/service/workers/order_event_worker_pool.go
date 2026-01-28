package workers

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type orderEventWorkerPool struct {
	workerCtx  *orderEventWorkerContext
	poolSize   uint
	wg         sync.WaitGroup
	outgoingCh chan<- OrderEvent
}

func newOrderEventWorkerPool(
	orderRepository OrderRepository,
	accrualClient AccrualClient,
) *orderEventWorkerPool {
	workerCtx := &orderEventWorkerContext{
		orderRepo:       orderRepository,
		accrualClient:   accrualClient,
		perEventTimeout: orderProcessorDefaultPerEventTimeout,
		incomingCh:      nil,
		callbackCh:      nil,
	}

	return &orderEventWorkerPool{
		workerCtx:  workerCtx,
		poolSize:   orderProcessorDefaultWorkerPoolSize,
		wg:         sync.WaitGroup{},
		outgoingCh: nil,
	}
}

// SetPerEventTimeout updates value of per-event processing timeout.
func (p *orderEventWorkerPool) SetPerEventTimeout(timeout time.Duration) {
	p.workerCtx.perEventTimeout = timeout
}

// SetPoolSize updates value of the pool size, that is
// the number of workers to be launched when [Start] is invoked.
func (p *orderEventWorkerPool) SetPoolSize(size uint) {
	if size == 0 {
		size = 1
	}

	p.poolSize = size
}

// Start launches N workers in goroutines, where N corresponds
// to a pre-configured pool size.
func (p *orderEventWorkerPool) Start(ctx context.Context, callbackCh chan<- orderEventResult) {
	workCh := make(chan OrderEvent)

	p.outgoingCh = workCh
	p.workerCtx.incomingCh = workCh
	p.workerCtx.callbackCh = callbackCh

	for id := range p.poolSize {
		p.startWorker(ctx, id)
	}

	slog.DebugContext(ctx, "pool: workers started")
}

// Wait waits for all workers to finish their work.
// Each worker would respect context cancellation and finish
// its work as soon as possible.
func (p *orderEventWorkerPool) Wait() {
	close(p.outgoingCh)

	p.wg.Wait()

	close(p.workerCtx.callbackCh)

	slog.Debug("pool: workers finished")
}

// TakeEvent takes provided event and sends it to the outgoing channel,
// where it would be picked up by one of the idle workers.
func (p *orderEventWorkerPool) TakeEvent(event OrderEvent) {
	p.outgoingCh <- event

	slog.Debug("pool: accepted for processing",
		slog.Any("event", event),
	)
}

func (p *orderEventWorkerPool) startWorker(
	ctx context.Context,
	workerID uint,
) {
	worker := orderEventWorker{
		orderEventWorkerContext: p.workerCtx,
		logger:                  slog.With(slog.Uint64("worker_id", uint64(workerID))),
	}

	p.wg.Add(1)

	go func() {
		defer p.wg.Done()

		worker.Run(ctx)
	}()
}
