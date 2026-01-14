package service

import "errors"

var (
	// ErrTokenNotValid is returned by [TokenManager.ValidateToken] when parsed token
	// is not considered valid by [github.com/golang-jwt/jwt] package.
	ErrTokenNotValid = errors.New("JWT token is not valid")

	// ErrTokenInvalidUserID is returned by [TokenManager.ValidateToken] when parsed token
	// contains empty or otherwise malformed user ID.
	ErrTokenInvalidUserID = errors.New("JWT token contains invalid user ID")
)
