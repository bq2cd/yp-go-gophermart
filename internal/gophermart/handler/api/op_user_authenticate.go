package api

import (
	"net/http"
)

// OperationAuthenticateUser is responsible for authenticating existing users.
// Available to anyone (public endpoint).
type OperationAuthenticateUser struct {
	operationCtxRegisterOrAuthenticate
}

// RespondUnauthorized returns '401 Unauthorized' when user provides
// incorrect [LoginPassword] credentials.
func (op *OperationAuthenticateUser) RespondUnauthorized() {
	op.ginCtx.AbortWithStatus(http.StatusUnauthorized)
}
