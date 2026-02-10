package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// Ensure [Handler] implements [api.Handler].
var _ api.Handler = (*Handler)(nil)

// Handler provides implementation for [api.Handler] operations.
type Handler struct {
	tokenService   TokenService
	userService    UserService
	balanceService BalanceService
	orderService   OrderService
}

// NewHandler initializes [Handler].
func NewHandler(
	tokenService TokenService,
	userService UserService,
	balanceService BalanceService,
	orderService OrderService,
) *Handler {
	return &Handler{
		tokenService:   tokenService,
		userService:    userService,
		balanceService: balanceService,
		orderService:   orderService,
	}
}
