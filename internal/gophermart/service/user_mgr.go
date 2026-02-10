package service

import (
	"context"
	"fmt"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"golang.org/x/crypto/bcrypt"
)

// UserManager implements [handler.UserService] interface.
type UserManager struct {
	userRepo   UserRepository
	bcryptCost int
}

// NewUserManager creates an instance of [UserManager].
func NewUserManager(
	userRepository UserRepository,
) *UserManager {
	return &UserManager{
		userRepo:   userRepository,
		bcryptCost: bcrypt.DefaultCost,
	}
}

// Register performs the creation of a new user with given user ID
// and plain text password.
// It returns [domain.ErrUserIDConflict] if another user with the same
// user ID already exists in the system.
// Other errors could also be returned.
func (m *UserManager) Register(ctx context.Context, userID domain.UserID, passwordPlain domain.PasswordPlain) error {
	passwordHash, err := m.GeneratePasswordHash(passwordPlain)
	if err != nil {
		return fmt.Errorf("cannot generate password hash: %w", err)
	}

	return m.createUser(ctx, userID, passwordHash)
}

// Authenticate checks if user ID already exists in the system
// and has a matching password.
// It returns [domain.ErrUserAuthenticationFailed] if user ID does not
// exist or provided password does not match the stored one.
// Other errors could also be returned.
func (m *UserManager) Authenticate(
	ctx context.Context,
	userID domain.UserID,
	passwordPlain domain.PasswordPlain,
) error {
	passwordHash, err := m.userRepo.GetPasswordHash(ctx, userID)
	if err != nil {
		return fmt.Errorf("cannot get user password: %w", err)
	}

	err = m.ValidatePasswordHash(passwordPlain, passwordHash)
	if err != nil {
		return domain.ErrUserAuthenticationFailed
	}

	return nil
}

// GeneratePasswordHash generates [bcrypt] hash from a plaintext password.
// An error is returned if the hash generation fails.
func (m *UserManager) GeneratePasswordHash(passwordPlain domain.PasswordPlain) (domain.PasswordHash, error) {
	if passwordPlain.IsEmpty() {
		return nil, domain.ErrEmptyPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(passwordPlain), m.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt failed to generate password hash: %w", err)
	}

	return domain.PasswordHash(hash), nil
}

// ValidatePasswordHash calls [bcrypt] to perform validation of a provided password hash and a plaintext password.
// It returns error if the provided password hash cannot be generated
// from the provided plaintext password.
func (m *UserManager) ValidatePasswordHash(passwordPlain domain.PasswordPlain, passwordHash domain.PasswordHash) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(passwordPlain))
	if err != nil {
		return fmt.Errorf("bcrypt failed to compare hash and password: %w", err)
	}

	return nil
}

func (m *UserManager) createUser(ctx context.Context, userID domain.UserID, passwordHash domain.PasswordHash) error {
	created, err := m.userRepo.CreateUser(ctx, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("cannot create user: %w", err)
	}

	if !created {
		return domain.ErrUserIDConflict
	}

	return nil
}
