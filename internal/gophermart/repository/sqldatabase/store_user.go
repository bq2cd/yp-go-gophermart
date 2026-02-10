package sqldatabase

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

// CreateUser add a new user record into the storage.
// It will return [domain.ErrUserIDConflict] if another user exists
// with the same user ID.
func (s *Storage) CreateUser(
	ctx context.Context,
	userID domain.UserID,
	passwordHash domain.PasswordHash,
) (bool, error) {
	_, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return false, err
	}

	if exists {
		return false, domain.ErrUserIDConflict
	}

	err = transactionCreateUserAndBalance(ctx, userID, passwordHash).
		Run(s)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetPasswordHash returns password hash bytes for an existing user.
// If the user does not exist, it will return [nil] password hash
// and no error.
// An error will only be returned in case of problems with
// the internal storage.
func (s *Storage) GetPasswordHash(ctx context.Context, userID domain.UserID) (domain.PasswordHash, error) {
	user, exists, err := s.findUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	return domain.PasswordHash(user.PasswordHash), nil
}

// findUser returns populated [models.User] and [true] if user exists.
// If user does not exist, it will return empty model and [false].
// An error will be returned if there were problems with the database.
func (s *Storage) findUser(ctx context.Context, userID domain.UserID) (models.User, bool, error) {
	user, err := s.getUser(ctx, userID)

	switch {
	case err == nil:
		return user, true, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return user, false, nil
	default:
		return user, false, err
	}
}

func (s *Storage) getUser(ctx context.Context, userID domain.UserID) (models.User, error) {
	user, err := Query[models.User](s).
		Where(generated.User.Login.Eq(userID.String())).
		First(ctx)
	if err != nil {
		return user, fmt.Errorf("cannot search users: %w", err)
	}

	return user, nil
}
