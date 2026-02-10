package app

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// App represent main app process consisting of several threads.
type App struct {
	threads []Thread
}

// NewApp creates an instance of [App].
func NewApp(threads ...Thread) *App {
	return &App{
		threads: threads,
	}
}

// Run launches main app's process which starts all app's threads
// and waits either for their completion (either successful or with an error).
// Threads are expected to shut down gracefully when provided context
// is canceled.
func (a *App) Run(ctx context.Context) error {
	grp := a.startThreads(ctx)

	return a.waitForShutdown(grp)
}

func (a *App) startThreads(baseCtx context.Context) *errgroup.Group {
	grp, ctx := errgroup.WithContext(baseCtx)

	for _, thread := range a.threads {
		grp.Go(func() error {
			return thread.Run(ctx)
		})
	}

	return grp
}

func (a *App) waitForShutdown(grp *errgroup.Group) error {
	err := grp.Wait()
	if err != nil {
		return fmt.Errorf("app failed: %w", err)
	}

	return nil
}
