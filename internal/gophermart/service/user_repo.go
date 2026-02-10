package service

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// UserRepository is responsible for storing/updating/retrieving information
// about users.
//
//go:generate mise run mockgen --outfile=user_repository.go UserRepository
type UserRepository interface {
	CreateUser(ctx context.Context, userID domain.UserID, passwordHash domain.PasswordHash) (bool, error)
	GetPasswordHash(ctx context.Context, userID domain.UserID) (domain.PasswordHash, error)
}
