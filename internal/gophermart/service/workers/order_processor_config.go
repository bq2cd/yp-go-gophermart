package workers

import (
	"time"
)

type orderProcessorConfig struct {
	ShutdownTimeout              time.Duration
	DelayConfig                  OrderEventDelayConfig
	EnableOrderPreloadingOnStart bool
}
