package workers

import "errors"

var (
	// ErrOrderEventDiscarded is an internal error for [OrderProcessor]to indicate that an event should be discarded.
	ErrOrderEventDiscarded = errors.New("event is not processable")
)
