package workers

import (
	"time"

	"github.com/avast/retry-go/v5"
)

type orderProcessorConfig struct {
	ShutdownTimeout              time.Duration
	DelayConfig                  retry.DelayContext
	EnableOrderPreloadingOnStart bool
}
