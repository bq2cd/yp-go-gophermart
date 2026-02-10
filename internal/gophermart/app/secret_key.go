package app

import (
	"context"
	"crypto/rand"
)

const (
	secretKeyDefaultRandomSizeInBytes = 32
)

// StaticSecretKey holds secret key bytes in memory.
// It implements [service.SecretKeyProvider] interface.
type StaticSecretKey struct {
	key []byte
}

// NewStaticSecretKey creates an instance [NewStaticSecretKey].
func NewStaticSecretKey(key []byte) *StaticSecretKey {
	return &StaticSecretKey{
		key: key,
	}
}

// GetSecretKey returns secret key bytes from memory, so
// it never returns an error.
// This method is needed for implementation of [service.SecretKeyProvider].
func (sk *StaticSecretKey) GetSecretKey(_ context.Context) ([]byte, error) {
	return sk.key, nil
}

// MustGenerateRandomSecretKey creates a new secret key using [crypto/rand]
// random generator.
// It will panic if [crypto/rand] module returns an error,
// which should never happen in practice, as per [crypto/rand] documentation.
func MustGenerateRandomSecretKey() []byte {
	buf := make([]byte, secretKeyDefaultRandomSizeInBytes)

	_, err := rand.Read(buf)
	if err != nil {
		panic("cannot generate random bytes")
	}

	return buf
}
