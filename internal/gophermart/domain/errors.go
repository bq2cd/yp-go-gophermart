package domain

import "errors"

var (
	// ErrUserIDConflict is returned during user registration process if a user already
	// exists with given [UserID].
	ErrUserIDConflict = errors.New("user with such ID already exists")

	// ErrUserAuthenticationFailed is returned during user authentication process when
	// there is either a credentials mismatch or user does not exist.
	ErrUserAuthenticationFailed = errors.New("user authentication failed")
)
