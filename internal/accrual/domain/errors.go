package domain

import "errors"

var (
	// ErrOrderNotFound is returned by an accrual system client when
	// the system responds with '204 No Content' HTTP status.
	ErrOrderNotFound = errors.New("order with such ID does not exist")

	// ErrRateLimitExceeded is returned by an accrual system client when
	// the system responds with '429 Too Many Requests' HTTP status.
	ErrRateLimitExceeded = errors.New("request rate limit exceeded")
)
