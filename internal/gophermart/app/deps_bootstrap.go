package app

import (
	"fmt"

	"github.com/bq2cd/yp-go-gophermart/internal/accrual/repository/apiclient"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
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
func BuildBootstrapDeps(config ConfigBootstrap) (BootstrapDeps, error) {
	builder := &bootstrapBuilder{
		config: config,
	}

	return builder.Build()
}

type bootstrapBuilder struct {
	config ConfigBootstrap
}

func (b *bootstrapBuilder) Build() (BootstrapDeps, error) {
	var deps BootstrapDeps

	storage, err := b.buildStorage()
	if err != nil {
		return deps, fmt.Errorf("cannot build storage: %w", err)
	}

	deps.Storage = storage
	deps.AccrualClient = b.buildAccrualClient()
	deps.SecretKeyProvider = b.buildSecretKeyProvider()

	return deps, nil
}

//nolint:ireturn
func (b *bootstrapBuilder) buildStorage() (Storage, error) {
	storage, err := sqldatabase.NewStorage(b.config.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize SQL database: %w", err)
	}

	err = storage.AutoMigrate()
	if err != nil {
		return nil, fmt.Errorf("cannot apply migrations: %w", err)
	}

	return storage, nil
}

func (b *bootstrapBuilder) buildAccrualClient() *apiclient.Client {
	return apiclient.NewClient(b.config.AccrualSystemURL)
}

func (b *bootstrapBuilder) buildSecretKeyProvider() *StaticSecretKey {
	return NewStaticSecretKey(b.config.AuthTokenSecretKey)
}
