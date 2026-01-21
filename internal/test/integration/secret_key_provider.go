package integration

import (
	"context"
	"crypto/rand"
)

type TestSecretKeyProvider struct {
	secretKey []byte
}

func NewTestSecretKeyProvider() *TestSecretKeyProvider {
	provider := &TestSecretKeyProvider{
		secretKey: make([]byte, 32),
	}

	provider.initSecretKey()

	return provider
}

func (p *TestSecretKeyProvider) GetSecretKey(ctx context.Context) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return p.secretKey, nil
}

func (p *TestSecretKeyProvider) initSecretKey() {
	// This never returns an error as per documentation.
	//nolint:errcheck,gosec
	rand.Read(p.secretKey)
}
