package app

import (
	"time"
)

// ConfigBootstrap contains configuration options for [BootstrapDeps].
type ConfigBootstrap struct {
	AccrualSystemURL   string
	AuthTokenSecretKey []byte
	DatabaseURI        string
}

// ConfigRuntime contains configuration options for [Runtime].
type ConfigRuntime struct {
	AuthTokenLifetime            time.Duration
	HTTPListenAddress            string
	HTTPShutdownTimeout          time.Duration
	OrderProcessorWorkerPoolSize uint
}
