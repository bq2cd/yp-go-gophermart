package workers

import (
	"math"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

const (
	orderEventDelayMaxBackOffN            uint = 62
	orderEventDelayInitialDelay                = 100 * time.Millisecond
	orderEventDelayMaxJitter                   = 100 * time.Millisecond
	orderEventDelayMinRetriesUntilBackOff      = 2
)

var _ retry.DelayContext = (*OrderEventDelayConfig)(nil)

// OrderEventDelayConfig implements [retry.DelayContext] interface.
type OrderEventDelayConfig struct {
	initialDelay           time.Duration
	maxJitter              time.Duration
	maxDelay               time.Duration
	maxBackOffN            uint
	minRetriesUntilBackOff uint
}

// NewOrderEventDelayConfig creates an instance of [OrderEventDelayConfig].
// By default, it sets initial delay and max jitter to 100 milliseconds,
// and max delay to 0 (which means unlimited retries).
func NewOrderEventDelayConfig(opts ...option.Option[OrderEventDelayConfig]) OrderEventDelayConfig {
	config := OrderEventDelayConfig{
		initialDelay:           orderEventDelayInitialDelay,
		maxJitter:              orderEventDelayMaxJitter,
		maxDelay:               0,
		maxBackOffN:            0,
		minRetriesUntilBackOff: orderEventDelayMinRetriesUntilBackOff,
	}

	for _, opt := range opts {
		opt(&config)
	}

	// Shameless copy-n-paste from retry-go internals.
	config.maxBackOffN = orderEventDelayMaxBackOffN - uint(math.Floor(math.Log2(float64(config.initialDelay))))

	return config
}

// WithOrderEventInitialDelay returns an option to configure initial delay for [OrderEventDelayConfig].
func WithOrderEventInitialDelay(delay time.Duration) option.Option[OrderEventDelayConfig] {
	return func(c *OrderEventDelayConfig) {
		c.initialDelay = delay
	}
}

// WithOrderEventMaxJitter returns an option to configure maximum jitter for [OrderEventDelayConfig].
func WithOrderEventMaxJitter(jitter time.Duration) option.Option[OrderEventDelayConfig] {
	return func(c *OrderEventDelayConfig) {
		c.maxJitter = jitter
	}
}

// WithOrderEventMaxDelay returns an option to configure maximum delay for [OrderEventDelayConfig].
func WithOrderEventMaxDelay(delay time.Duration) option.Option[OrderEventDelayConfig] {
	return func(c *OrderEventDelayConfig) {
		c.maxDelay = delay
	}
}

// WithOrderEventMinRetriesUntilBackOff returns an option to configure [OrderEventDelayConfig] with
// minimum retry attempts until [retry.BackOffDelay] strategy kicks in.
// Until that many attempts have passed, a simpler [retry.RandomDelay] strategy will be at work.
// This method is exposed primarily for testing purposes.
func WithOrderEventMinRetriesUntilBackOff(retries uint) option.Option[OrderEventDelayConfig] {
	return func(c *OrderEventDelayConfig) {
		c.minRetriesUntilBackOff = retries
	}
}

// Delay returns initial delay, as per [retry.DelayContext] contract.
func (c *OrderEventDelayConfig) Delay() time.Duration {
	return c.initialDelay
}

// MaxJitter returns maximum jitter, as per [retry.DelayContext] contract.
func (c *OrderEventDelayConfig) MaxJitter() time.Duration {
	return c.maxJitter
}

// MaxDelay returns maximum delay, as per [retry.DelayContext] contract.
func (c *OrderEventDelayConfig) MaxDelay() time.Duration {
	return c.maxDelay
}

// MaxBackOffN returns maximum back-off factor, as per [retry.DelayContext] contract.
func (c *OrderEventDelayConfig) MaxBackOffN() uint {
	return c.maxBackOffN
}

// MinRetriesUntilBackOff returns minimum retry attempts needed to activate [retry.BackOffDelay] strategy.
// Until that many attempts have passed, a simpler [retry.RandomDelay] strategy will be at work.
func (c *OrderEventDelayConfig) MinRetriesUntilBackOff() uint {
	return c.minRetriesUntilBackOff
}
