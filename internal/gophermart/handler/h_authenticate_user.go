package handler

import (
	"errors"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// AuthenticateUser implements [api.OperationAuthenticateUser].
func (h *Handler) AuthenticateUser(operation *api.OperationAuthenticateUser) {
	userID, passwordPlain, ok := processUserRegistrationOrAuthenticationRequest(operation)
	if !ok {
		return
	}

	err := h.userService.Authenticate(operation.Context(), userID, passwordPlain)
	if err != nil {
		h.processAuthenticateUserError(operation, err)

		return
	}

	issueUserToken(h.tokenService, operation, userID)
}

func (h *Handler) processAuthenticateUserError(operation *api.OperationAuthenticateUser, err error) {
	if errors.Is(err, domain.ErrUserAuthenticationFailed) {
		operation.RespondUnauthorized()

		return
	}

	operation.RespondServerError()
}
