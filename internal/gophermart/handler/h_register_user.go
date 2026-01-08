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

	h.issueToken(operation, userID)
}

func (h *Handler) sendRegisterUserError(operation *api.OperationRegisterUser, err error) {
	if errors.Is(err, domain.ErrUserIDConflict) {
		operation.RespondConflict()

		return
	}

	operation.RespondServerError()
}

func (h *Handler) issueToken(operation *api.OperationRegisterUser, userID domain.UserID) {
	token, err := h.tokenService.IssueToken(userID)
	if err != nil {
		operation.RespondServerError()

		return
	}

	operation.RespondOK(api.UserAuthenticated{Token: token.String()})
}
