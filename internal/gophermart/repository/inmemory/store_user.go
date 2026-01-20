package inmemory

import (
	"context"
	"fmt"

	"github.com/govalues/decimal"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

// CreateUser add a new user record into the storage.
// It will return [domain.ErrUserIDConflict] if another user exists
// with the same user ID.
func (s *Storage) CreateUser(
	_ context.Context,
	userID domain.UserID,
	passwordHash domain.PasswordHash,
) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.users[userID]
	if ok {
		return false, nil
	}

	s.users[userID] = User{
		Password:             passwordHash,
		Balance:              decimal.Zero,
		TotalAmountWithdrawn: decimal.Zero,
	}

	return true, nil
}

// GetPasswordHash returns password hash bytes for an existing user.
// If the user does not exist, it will return [nil] password hash
// and no error.
// An error will only be returned in case of problems with
// the internal storage.
func (s *Storage) GetPasswordHash(_ context.Context, userID domain.UserID) (domain.PasswordHash, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return nil, nil
	}

	return user.Password, nil
}

// updateUser is intended to be executed atomically.
// It is the responsibility of a caller to ensure that this call is
// protected from race conditions.
func (s *Storage) updateUser(userID domain.UserID, updateFn UserUpdateFn) (bool, error) {
	user, ok := s.users[userID]
	if !ok {
		return false, domain.ErrUserNotFound
	}

	success, err := updateFn(&user)
	if err != nil {
		return false, fmt.Errorf("cannot update user: %w", err)
	}

	if !success {
		return false, nil
	}

	s.users[userID] = user

	return true, nil
}
