package handler

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

type userAuthenticatedOperation interface {
	RespondServerError()
	RespondOK(resp api.UserAuthenticated)
}

func issueUserToken(tokenService TokenService, operation userAuthenticatedOperation, userID domain.UserID) {
	token, err := tokenService.IssueToken(userID)
	if err != nil {
		operation.RespondServerError()

		return
	}

	operation.RespondOK(api.UserAuthenticated{Token: token.String()})
}

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
