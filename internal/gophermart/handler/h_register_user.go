package handler

import (
	"errors"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// RegisterUser implements [api.OperationRegisterUser].
func (h *Handler) RegisterUser(operation *api.OperationRegisterUser) {
	req, err := operation.GetRequest()
	if err != nil {
		operation.RespondBadRequest()

		return
	}

	userID := domain.UserID(req.Login)

	err = h.userService.Register(userID, domain.PasswordPlain(req.Password))
	if err != nil {
		h.sendRegisterUserError(operation, err)

		return
	}

	issueUserToken(h.tokenService, operation, userID)
}

func (h *Handler) sendRegisterUserError(operation *api.OperationRegisterUser, err error) {
	if errors.Is(err, domain.ErrUserIDConflict) {
		operation.RespondConflict()

		return
	}

	operation.RespondServerError()
}
