package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// Withdraw implements [api.OperationWithdraw].
func (h *Handler) Withdraw(operation *api.OperationWithdraw) {
	operation.RespondServerError()
}
