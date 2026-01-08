package handler

import "github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"

// UploadOrder implements [api.OperationUploadOrder].
func (h *Handler) UploadOrder(operation *api.OperationUploadOrder) {
	operation.RespondServerError()
}
