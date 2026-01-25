package app

import "time"

// ConfigBootstrap contains configuration options for [BootstrapDeps].
type ConfigBootstrap struct {
	AccrualSystemURL   string
	AuthTokenSecretKey []byte
}

// ConfigRuntime contains configuration options for [Runtime].
type ConfigRuntime struct {
	AuthTokenLifetime   time.Duration
	HTTPListenAddress   string
	HTTPShutdownTimeout time.Duration
}
