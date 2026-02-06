package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

type orderEventWorkerContext struct {
	orderRepo       OrderRepository
	accrualClient   AccrualClient
	perEventTimeout time.Duration
	incomingCh      <-chan OrderEvent
	callbackCh      chan<- orderEventResult
	mu              sync.RWMutex
	pausedUntil     time.Time
}

type orderEventWorker struct {
	*orderEventWorkerContext

	logger *slog.Logger
}

// Run launches main loop of a worker that consumes events from the incoming channel and processes them one by one.
func (w *orderEventWorker) Run(ctx context.Context) {
	w.mainLoop(ctx)

	w.logger.DebugContext(ctx, "worker: main loop completed")
}

func (w *orderEventWorker) mainLoop(ctx context.Context) {
	for event := range w.incomingCh {
		w.handleIncoming(ctx, event)
	}
}

func (w *orderEventWorker) handleIncoming(ctx context.Context, event OrderEvent) {
	var err error

	if !hasContextExpired(ctx) {
		err = w.processEvent(ctx, event)
	}

	w.callbackOnDone(event, err)
}

func (w *orderEventWorker) processEvent(baseCtx context.Context, event OrderEvent) error {
	ctx, cancel := context.WithTimeout(baseCtx, w.perEventTimeout)
	defer cancel()

	w.logger.DebugContext(ctx, "worker: prepare for processing",
		slog.Any("event", event),
		slog.Any("ctx", ctx),
	)

	shouldProcess, err := w.shouldProcessEvent(ctx, event)
	if err != nil {
		return err
	}

	if !shouldProcess {
		return nil
	}

	accOrder, err := w.makeAccrualClientRequest(ctx, event)
	if err != nil {
		return err
	}

	return w.processAccrualClientOrder(ctx, event.UserID, accOrder)
}

func (w *orderEventWorker) shouldProcessEvent(ctx context.Context, event OrderEvent) (bool, error) {
	status, err := w.orderRepo.GetOrderStatus(ctx, event.UserID, event.OrderID)
	if err != nil {
		return false, fmt.Errorf("cannot get order status: %w", err)
	}

	return w.shouldProcessOrderStatus(ctx, event, status), nil
}

func (w *orderEventWorker) shouldProcessOrderStatus(
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

	w.logger.DebugContext(ctx, "worker: verify eligibility for processing",
		slog.Group("order",
			slog.Uint64("id", uint64(event.OrderID)),
			slog.String("user", string(event.UserID)),
			slog.Int("status", int(status)),
		),
		slog.Bool("is_eligible", isEligible),
	)

	return isEligible
}

func (w *orderEventWorker) makeAccrualClientRequest(ctx context.Context, event OrderEvent) (accdomain.Order, error) {
	var (
		accOrder       accdomain.Order
		err            error
		rateLimitError *accdomain.RateLimitExceededError
	)

	w.waitUntilUnpaused(ctx)

	if hasContextExpired(ctx) {
		return accOrder, fmt.Errorf("cannot send request to accrual system: %w", ctx.Err())
	}

	accOrder, err = w.accrualClient.GetOrderStatus(ctx, accdomain.OrderID(event.OrderID))

	switch {
	case err == nil:
		return accOrder, nil
	case errors.As(err, &rateLimitError):
		w.setPausedUntil(time.Now().Add(rateLimitError.RetryAfter))
	}

	return accOrder, fmt.Errorf("cannot get order status from accrual system: %w", err)
}

func (w *orderEventWorker) processAccrualClientOrder(
	ctx context.Context,
	userID domain.UserID,
	accOrder accdomain.Order,
) error {
	var err error

	orderID := domain.OrderID(accOrder.ID)

	switch accOrder.Status {
	case accdomain.OrderStatusRegistered, accdomain.OrderStatusProcessing:
		err = w.orderRepo.MarkOrderProcessing(ctx, userID, orderID)
	case accdomain.OrderStatusProcessed:
		err = w.orderRepo.MarkOrderProcessed(ctx, userID, orderID, accOrder.AccrualPoints)
	case accdomain.OrderStatusInvalid:
		err = w.orderRepo.MarkOrderInvalid(ctx, userID, orderID)
	default:
		err = w.orderRepo.MarkOrderInvalid(ctx, userID, orderID)
	}

	if err != nil {
		return fmt.Errorf("cannot change order status: %w", err)
	}

	return nil
}

func (w *orderEventWorker) callbackOnDone(event OrderEvent, err error) {
	w.logger.Debug("worker: sending result",
		slog.Any("event", event),
		slog.Any("error", err),
	)

	w.callbackCh <- orderEventResult{
		OrderEvent: event,
		err:        err,
	}
}

func (w *orderEventWorker) setPausedUntil(timestamp time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.pausedUntil = timestamp
}

func (w *orderEventWorker) getPausedUntil() time.Time {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.pausedUntil
}

func (w *orderEventWorker) waitUntilUnpaused(ctx context.Context) {
	timer := time.NewTimer(time.Until(w.getPausedUntil()))

	select {
	case <-ctx.Done():
		timer.Stop()
	case <-timer.C:
	}
}
