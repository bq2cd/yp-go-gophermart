package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// ListOrders implements [api.OperationListOrders].
func (h *Handler) ListOrders(operation *api.OperationListOrders) {
	operation.RespondServerError()
}
