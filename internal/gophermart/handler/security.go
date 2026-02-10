package handler

import (
	"context"
	"fmt"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// Ensure [SecurityHandler] implements [api.SecurityHandler].
var _ api.SecurityHandler = (*SecurityHandler)(nil)

// SecurityHandler provides implementation for [api.SecurityHandler] interface.
type SecurityHandler struct {
	tokenService TokenService
}

// NewSecurityHandler initializes [SecurityHandler].
func NewSecurityHandler(tokenService TokenService) *SecurityHandler {
	return &SecurityHandler{
		tokenService: tokenService,
	}
}

// JWTAuth performs authentication based on provided JWT token.
func (sh *SecurityHandler) JWTAuth(ctx context.Context, token string) (api.UserID, error) {
	userID, err := sh.tokenService.ValidateToken(ctx, domain.Token(token))
	if err != nil {
		return api.UserIDEmptyValue, fmt.Errorf("%w: %w", ErrAuthTokenInvalid, err)
	}

	return api.UserID(userID), nil
}
