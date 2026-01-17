package workers

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderItem represent an internal piece of work for the [OrderProcessor].
type OrderItem struct {
	UserID  domain.UserID
	OrderID domain.OrderID
}

// OrderQueue defines an in-memory queue as a buffer for
// incoming [OrderItem] items.
// This queue is assumed to be not thread-safe, so it needs
// to be protected with locks.
type OrderQueue interface {
	PushBack(item OrderItem)
	// PopFront panics on empty queue.
	// Use [Len] method to check if queue is empty.
	PopFront() OrderItem
	Len() int
}
