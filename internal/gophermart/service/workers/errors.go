package workers

import "errors"

var (
	// ErrOrderEventDiscarded is an internal error for [OrderProcessor] to indicate that an event should be discarded.
	ErrOrderEventDiscarded = errors.New("event is not processable")

	// ErrOrderIsNotReady is an internal error for [OrderProcessor] to indicate that order processing should be
	// postponed to avoid hammering the accrual system.
	ErrOrderIsNotReady = errors.New("order has not been processed by the accrual system yet")
)
