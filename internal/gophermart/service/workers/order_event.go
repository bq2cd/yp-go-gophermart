package workers

import (
	"errors"
	"log/slog"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderEvent represent an internal event for the [OrderProcessor]
// used during processing.
type OrderEvent struct {
	userID       domain.UserID
	orderID      domain.OrderID
	processAfter time.Time
	retries      uint
	delayConfig  retry.DelayContext
}

// LogValue implements [slog.LogValuer] interface to render [OrderEvent] in logs.
func (ev OrderEvent) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("user_id", ev.orderID.String()),
		slog.Uint64("order_id", uint64(ev.orderID)),
		slog.Time("process_after", ev.processAfter),
		slog.Uint64("retries", uint64(ev.retries)),
	)
}

type orderEventResult struct {
	OrderEvent

	err error
}

// GetRetryableEvent will return either original event and [false] if it cannot be retried, or
// a modified event and [true] if such event can be retried.
func (r *orderEventResult) GetRetryableEvent() (OrderEvent, bool) {
	return r.getNextEvent()
}

func (r *orderEventResult) getNextEvent() (OrderEvent, bool) {
	retries := r.retries + 1

	processAfter, ok := r.getNextProcessingTime(retries)
	if !ok {
		return r.OrderEvent, false
	}

	next := OrderEvent{
		userID:       r.userID,
		orderID:      r.orderID,
		processAfter: processAfter,
		retries:      retries,
		delayConfig:  r.delayConfig,
	}

	return next, true
}

func (r *orderEventResult) getNextProcessingTime(retries uint) (time.Time, bool) {
	var zero time.Time

	delayFn, ok := r.getDelayFn(retries)
	if !ok {
		return zero, false
	}

	delay := delayFn(retries, r.err, r.delayConfig)
	maxDelay := r.delayConfig.MaxDelay()

	if maxDelay > 0 && delay > maxDelay {
		delay = maxDelay
	}

	next := time.Now().Add(delay)

	return next, true
}

func (r *orderEventResult) getDelayFn(retries uint) (retry.DelayTypeFunc, bool) {
	switch {
	case r.err == nil:
		return nil, false
	case errors.Is(r.err, ErrOrderEventDiscarded):
		return nil, false
	case errors.Is(r.err, domain.ErrOrderNotFound):
		return nil, false
	}

	if retries < orderEventDelayMinRetriesUntilBackOff {
		return retry.RandomDelay, true
	}

	delayFn := retry.CombineDelay(
		retry.BackOffDelay,
		retry.RandomDelay,
	)

	return delayFn, true
}
