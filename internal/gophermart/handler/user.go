package handler

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// ///////////////////////////////////////////////////////////////////////////////
type userRegistrationOrAuthenticationOperation interface {
	GetRequest() (*api.LoginPassword, error)
	RespondBadRequest()
}

func processUserRegistrationOrAuthenticationRequest(
	operation userRegistrationOrAuthenticationOperation,
) (domain.UserID, domain.PasswordPlain, bool) {
	req, err := operation.GetRequest()
	if err != nil {
		operation.RespondBadRequest()

		return domain.UserIDEmptyValue, domain.PasswordPlain(""), false
	}

	return domain.UserID(req.Login), domain.PasswordPlain(req.Password), true
}

/////////////////////////////////////////////////////////////////////////////////

type userAuthenticatedOperation interface {
	RespondServerError()
	RespondOK(resp api.UserAuthenticated)
	Context() context.Context
}

func issueUserToken(tokenService TokenService, operation userAuthenticatedOperation, userID domain.UserID) {
	token, err := tokenService.IssueToken(operation.Context(), userID)
	if err != nil {
		operation.RespondServerError()

		return
	}

	operation.RespondOK(api.UserAuthenticated{Token: token.String()})
}

/////////////////////////////////////////////////////////////////////////////////

type userIDGetter interface {
	GetUserID() (api.UserID, bool)
	RespondUnauthorized()
}

func (h *Handler) ensureUserID(operation userIDGetter) (domain.UserID, bool) {
	apiUserID, ok := operation.GetUserID()
	if !ok {
		operation.RespondUnauthorized()

		return domain.UserIDEmptyValue, false
	}

	return domain.UserID(apiUserID), true
}
