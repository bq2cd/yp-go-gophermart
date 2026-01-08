package domain

import "errors"

var (
	// ErrUserIDConflict is returned during user registration process if a user already
	// exists with given [UserID].
	ErrUserIDConflict = errors.New("user with such ID already exists")
)
