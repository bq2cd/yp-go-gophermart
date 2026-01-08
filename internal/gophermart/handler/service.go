package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"

// UserService provides methods to register/authenticate a user.
//
//go:generate go tool mockgen -typed -destination=mocks/user_service.go -package=mocks . UserService
type UserService interface {
	Register(userID domain.UserID, passwordPlain domain.PasswordPlain) error
	Authenticate(userID domain.UserID, passwordPlain domain.PasswordPlain) error
}

// TokenService provides methods to issue tokens an authenticated user.
//
//go:generate go tool mockgen -typed -destination=mocks/token_service.go -package=mocks . TokenService
type TokenService interface {
	IssueToken(userID domain.UserID) (domain.Token, error)
}
