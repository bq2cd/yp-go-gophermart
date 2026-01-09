package handler

import "errors"

var (
	// ErrAuthTokenInvalid is returned from [SecurityHandler.JWTAuth]
	// when token value fails to pass validation.
	ErrAuthTokenInvalid = errors.New("auth token validation error")
)
