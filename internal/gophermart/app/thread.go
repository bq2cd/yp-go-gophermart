package app

import (
	"context"
)

// Thread represents a background process executing in a goroutine.
// The goroutine is launched by an external entity, not by the thread itself,
// although the thread might launch extra goroutines internally.
// The thread is expected to shut down when provided context is canceled.
type Thread interface {
	Run(ctx context.Context) error
}
