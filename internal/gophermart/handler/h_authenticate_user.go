package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// AuthenticateUser implements [api.OperationAuthenticateUser].
func (h *Handler) AuthenticateUser(operation *api.OperationAuthenticateUser) {
	operation.RespondServerError()
}
