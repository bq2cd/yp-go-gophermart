package handler

import (
	"errors"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// RegisterUser implements [api.OperationRegisterUser].
func (h *Handler) RegisterUser(operation *api.OperationRegisterUser) {
	userID, passwordPlain, ok := processUserRegistrationOrAuthenticationRequest(operation)
	if !ok {
		return
	}

	err := h.userService.Register(operation.Context(), userID, passwordPlain)
	if err != nil {
		h.processRegisterUserError(operation, err)

		return
	}

	issueUserToken(h.tokenService, operation, userID)
}

func (h *Handler) processRegisterUserError(operation *api.OperationRegisterUser, err error) {
	if errors.Is(err, domain.ErrUserIDConflict) {
		operation.RespondConflict()

		return
	}

	operation.RespondServerError()
}
