package app

import (
	"context"
	"log/slog"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

// OrderProcessor is a wrapper for [workers.OrderProcessor] to implement [Thread] interface.
type OrderProcessor struct {
	processor *workers.OrderProcessor
}

// Run starts [workers.OrderProcessor] and waits for its completion.
// When provided context is canceled, the order processor will
// shut down gracefully.
// This method never returns an error because [workers.OrderProcessor.Run]
// does not return anything.
// The method provides such signature to match [Thread] interface.
func (p *OrderProcessor) Run(ctx context.Context) error {
	slog.DebugContext(ctx, "starting order processor")

	p.processor.Run(ctx)

	return nil
}

// HasFinished exposes [workers.OrderProcessor.HasFinished] method,
// primarily for testing purposes.
func (p *OrderProcessor) HasFinished() bool {
	return p.processor.HasFinished()
}
