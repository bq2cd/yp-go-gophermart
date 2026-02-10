package api

import (
	"net/http"
)

// OperationRegisterUser is responsible for registering new users.
// Available to anyone (public endpoint).
type OperationRegisterUser struct {
	operationCtxRegisterOrAuthenticate
}

// RespondConflict returns '409 Conflict' when there is already existing user
// with the same login as provided in [LoginPassword].
func (op *OperationRegisterUser) RespondConflict() {
	op.ginCtx.AbortWithStatus(http.StatusConflict)
}
