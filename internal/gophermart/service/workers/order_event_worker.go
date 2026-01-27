package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

	accOrder, err := w.accrualClient.GetOrderStatus(ctx, accdomain.OrderID(event.orderID))
	if err != nil {
		return w.processAccrualClientError(err)
	}

	return w.processAccrualClientOrder(ctx, event.userID, accOrder)
}

func (w *orderEventWorker) shouldProcessEvent(ctx context.Context, event OrderEvent) (bool, error) {
	status, err := w.orderRepo.GetOrderStatus(ctx, event.userID, event.orderID)
	if err != nil {
		return false, w.processGetOrderStatusError(err)
	}

	return w.shouldProcessOrderStatus(ctx, event, status), nil
}

func (w *orderEventWorker) processGetOrderStatusError(err error) error {
	if errors.Is(err, domain.ErrOrderNotFound) {
		return nil
	}

	return err
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
			slog.Uint64("id", uint64(event.orderID)),
			slog.String("user", string(event.userID)),
			slog.Int("status", int(status)),
		),
		slog.Bool("is_eligible", isEligible),
	)

	return isEligible
}

func (w *orderEventWorker) processAccrualClientError(err error) error {
	return err
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
