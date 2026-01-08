package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// ListWithdrawals implements [api.OperationListWithdrawals].
func (h *Handler) ListWithdrawals(operation *api.OperationListWithdrawals) {
	operation.RespondServerError()
}
