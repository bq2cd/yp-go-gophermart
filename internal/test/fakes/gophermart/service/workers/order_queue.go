//nolint:revive,exhaustruct
package fakes

import (
	"slices"
	"sync"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

var _ workers.OrderQueue = (*TestOrderQueue)(nil)

type TestOrderQueue struct {
	mu    sync.RWMutex
	items []workers.OrderItem
}

func NewTestOrderQueue() *TestOrderQueue {
	return &TestOrderQueue{
		items: make([]workers.OrderItem, 0),
	}
}

func (q *TestOrderQueue) PushBack(item workers.OrderItem) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, item)
}

func (q *TestOrderQueue) PopFront() workers.OrderItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := q.items[0]
	q.items = q.items[1:]

	return out
}

func (q *TestOrderQueue) Len() int {
	q.mu.RLock()
	n := len(q.items)
	q.mu.RUnlock()

	return n
}

// CopyToSlice is for testing purposes.
func (q *TestOrderQueue) CopyToSlice() []workers.OrderItem {
	return slices.Clone(q.items)
}

// CopyFromSlice is for testing purposes.
func (q *TestOrderQueue) CopyFromSlice(items []workers.OrderItem) {
	q.items = slices.Clone(items)
}
