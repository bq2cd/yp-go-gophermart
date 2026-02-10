package workers

import (
	"context"
	"time"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// OrderQueue provides a way for [OrderProcessor] to obtain the next
// order that needs processing and to postpone processing of a given
// order until a later date (e.g. due to errors).
type OrderQueue interface {
	// GetNextProcessableOrder returns a [domain.ProcessableOrder] and a number of prior retries for it.
	GetNextProcessableOrder(ctx context.Context, excludeIDs []domain.OrderID) (domain.ProcessableOrder, uint, error)
	// PostponeOrderProcessing postpones given order processing to a later timestamp.
	// It will increment number of retries internally.
	PostponeOrderProcessing(ctx context.Context, order domain.ProcessableOrder, processAfter time.Time) error
}
