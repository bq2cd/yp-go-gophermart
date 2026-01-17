package service

import "context"

// SecretKeyProvider is responsible for exposing a secret key for signing JWT tokens.
//
//go:generate go tool mockgen -typed -destination=mocks/secret_key_provider.go -package=mocks . SecretKeyProvider
type SecretKeyProvider interface {
	GetSecretKey(ctx context.Context) ([]byte, error)
}
