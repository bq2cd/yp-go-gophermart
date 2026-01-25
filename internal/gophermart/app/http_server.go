package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

const (
	httpServerDefaultShutdownTimeout   = 5 * time.Second
	httpServerDefaultReadHeaderTimeout = 1 * time.Second
)

// HTTPServer wraps [http.Server] to provide additional
// life cycle management functions, such as shutdown.
type HTTPServer struct {
	server          *http.Server
	shutdownTimeout time.Duration
}

// WithHTTPServerShutdownTimeout returns an option to configure [HTTPServer] with provided
// timeout for graceful shutdown.
func WithHTTPServerShutdownTimeout(timeout time.Duration) option.Option[HTTPServer] {
	return func(srv *HTTPServer) {
		srv.shutdownTimeout = timeout
	}
}

// NewHTTPServer creates an instance of [HTTPServer].
func NewHTTPServer(addr string, handler http.Handler, opts ...option.Option[HTTPServer]) *HTTPServer {
	httpServer := &http.Server{ //nolint:exhaustruct
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: httpServerDefaultReadHeaderTimeout,
	}

	srv := &HTTPServer{
		server:          httpServer,
		shutdownTimeout: httpServerDefaultShutdownTimeout,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

// Run launches an HTTP server and monitors a provided context for
// cancellation.
// When the context is canceled, it will perform a graceful shutdown
// of the HTTP server.
func (s *HTTPServer) Run(ctx context.Context) error {
	errCh := s.startServer()

	return s.waitForShutdown(ctx, errCh)
}

func (s *HTTPServer) startServer() <-chan error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.listenAndServe()
	}()

	return errCh
}

func (s *HTTPServer) listenAndServe() error {
	err := s.server.ListenAndServe()

	switch {
	case err == nil:
		return nil
	case errors.Is(err, http.ErrServerClosed):
		return nil
	default:
		return fmt.Errorf("cannot start HTTP server: %w", err)
	}
}

func (s *HTTPServer) waitForShutdown(ctx context.Context, errCh <-chan error) error {
	for {
		select {
		case <-ctx.Done():
			return s.shutdown(context.WithoutCancel(ctx))
		case err := <-errCh:
			return err
		}
	}
}

func (s *HTTPServer) shutdown(baseCtx context.Context) error {
	ctx, cancel := context.WithTimeout(baseCtx, s.shutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("HTTP server failed to shutdown gracefully: %w", err)
	}

	return nil
}
