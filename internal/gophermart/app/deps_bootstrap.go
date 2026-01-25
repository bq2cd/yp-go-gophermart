package app

import (
	"github.com/bq2cd/yp-go-gophermart/internal/accrual/repository/apiclient"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/inmemory"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

// BootstrapDeps describes external and low-level dependencies
// that are needed to build [RuntimeDeps].
type BootstrapDeps struct {
	Storage           Storage
	AccrualClient     workers.AccrualClient
	SecretKeyProvider service.SecretKeyProvider
}

// BuildBootstrapDeps initializes [BootstrapDeps] dependencies
// using provided [ConfigBootstrap].
func BuildBootstrapDeps(config ConfigBootstrap) BootstrapDeps {
	builder := &bootstrapBuilder{
		config: config,
	}

	return builder.Build()
}

type bootstrapBuilder struct {
	config ConfigBootstrap
}

func (b *bootstrapBuilder) Build() BootstrapDeps {
	return BootstrapDeps{
		Storage:           b.buildStorage(),
		AccrualClient:     b.buildAccrualClient(),
		SecretKeyProvider: b.buildSecretKeyProvider(),
	}
}

func (b *bootstrapBuilder) buildStorage() *inmemory.Storage {
	return inmemory.NewStorage()
}

func (b *bootstrapBuilder) buildAccrualClient() *apiclient.Client {
	return apiclient.NewClient(b.config.AccrualSystemURL)
}

func (b *bootstrapBuilder) buildSecretKeyProvider() *StaticSecretKey {
	return NewStaticSecretKey(b.config.AuthTokenSecretKey)
}
