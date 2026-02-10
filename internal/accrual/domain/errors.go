package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrOrderNotFound is returned by an accrual system client when
	// the system responds with '204 No Content' HTTP status.
	ErrOrderNotFound = errors.New("order with such ID does not exist")
)

// RateLimitExceededError is returned by an accrual system client when
// the system responds with '429 Too Many Requests' HTTP status.
type RateLimitExceededError struct {
	RetryAfter time.Duration
}

// Error implements [error] interface for [RateLimitExceededError].
func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %v", e.RetryAfter)
}
