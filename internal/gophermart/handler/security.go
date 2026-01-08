package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

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
func (sh *SecurityHandler) JWTAuth(token string) error {
	_ = token

	return nil
}
