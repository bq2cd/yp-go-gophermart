package apiclient

import (
	"time"

	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

const (
	retryConfigDefaultCount       = 3
	retryConfigDefaultWaitTime    = 1 * time.Second
	retryConfigDefaultMaxWaitTime = 5 * time.Second
)

// RetryConfig groups together various settings for retries.
type RetryConfig struct {
	Count       int
	WaitTime    time.Duration
	MaxWaitTime time.Duration
}

// DefaultRetryConfig returns [RetryConfig] with default values.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		Count:       retryConfigDefaultCount,
		WaitTime:    retryConfigDefaultWaitTime,
		MaxWaitTime: retryConfigDefaultMaxWaitTime,
	}
}

// WithRetryConfig returns an option to override retry configuration
// of a [Client].
func WithRetryConfig(config RetryConfig) option.Option[Client] {
	return func(c *Client) {
		c.applyRetryConfig(config)
	}
}
