package service

import "context"

// SecretKeyProvider is responsible for exposing a secret key for signing JWT tokens.
//
//go:generate mise run mockgen --outfile=secret_key_provider.go SecretKeyProvider
type SecretKeyProvider interface {
	GetSecretKey(ctx context.Context) ([]byte, error)
}
