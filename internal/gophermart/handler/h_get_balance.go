package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// GetBalance implements [api.OperationGetBalance].
func (h *Handler) GetBalance(operation *api.OperationGetBalance) {
	operation.RespondServerError()
}
