package workers

import "github.com/gammazero/deque"

// OrderQueue defines an in-memory queue as a buffer for
// incoming [OrderEvent] events.
// This queue is assumed to be not thread-safe, so it needs
// to be protected with locks.
type OrderQueue interface {
	PushBack(event OrderEvent)
	// PopFront panics on empty queue.
	// Use [Len] method to check if queue is empty.
	PopFront() OrderEvent
	Len() int
}

// NewOrderQueue returns a concrete implementation of [OrderQueue] interface.
func NewOrderQueue() *deque.Deque[OrderEvent] {
	return new(deque.Deque[OrderEvent])
}
