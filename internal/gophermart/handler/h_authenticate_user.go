package handler

import (
	"errors"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// AuthenticateUser implements [api.OperationAuthenticateUser].
func (h *Handler) AuthenticateUser(operation *api.OperationAuthenticateUser) {
	req, err := operation.GetRequest()
	if err != nil {
		operation.RespondBadRequest()

		return
	}

	userID := domain.UserID(req.Login)

	err = h.userService.Authenticate(userID, domain.PasswordPlain(req.Password))
	if err != nil {
		h.sendAuthenticateUserError(operation, err)

		return
	}

	issueUserToken(h.tokenService, operation, userID)
}

func (h *Handler) sendAuthenticateUserError(operation *api.OperationAuthenticateUser, err error) {
	if errors.Is(err, domain.ErrUserAuthenticationFailed) {
		operation.RespondUnauthorized()

		return
	}

	operation.RespondServerError()
}
