package api

import (
	"net/http"
)

// LoginPassword defines user's credentials, both used for registering
// a new user and for authenticating an existing one.
type LoginPassword struct {
	Login    string `json:"login"    validate:"required,min=4"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

// UserAuthenticated defines server's response when user has been
// successfully registered or authenticated.
// It contains JWT token to be put into [AuthorizationHeaderName] with
// [AuthorizationHeaderValuePrefix] prefix.
type UserAuthenticated struct {
	Token string
}

type operationCtxRegisterOrAuthenticate struct {
	operationCtx
}

func (op *operationCtxRegisterOrAuthenticate) GetRequest() (*LoginPassword, error) {
	return operationGetRequest[LoginPassword](op.ginCtx)
}

func (op *operationCtxRegisterOrAuthenticate) RespondOK(resp UserAuthenticated) {
	op.ginCtx.Header(AuthorizationHeaderName, AuthorizationHeaderValuePrefix+resp.Token)
	op.ginCtx.Status(http.StatusOK)
}
