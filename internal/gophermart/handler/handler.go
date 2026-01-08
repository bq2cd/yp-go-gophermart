package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// Ensure [Handler] implements [api.Handler].
var _ api.Handler = (*Handler)(nil)

// Handler provides implementation for [api.Handler] operations.
type Handler struct {
	userService  UserService
	tokenService TokenService
}

// NewHandler initializes [Handler].
func NewHandler(userService UserService, tokenService TokenService) *Handler {
	return &Handler{
		userService:  userService,
		tokenService: tokenService,
	}
}
