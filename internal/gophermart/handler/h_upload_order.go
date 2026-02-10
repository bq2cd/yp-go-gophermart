package handler

import (
	"errors"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// UploadOrder implements [api.OperationUploadOrder].
func (h *Handler) UploadOrder(operation *api.OperationUploadOrder) {
	userID, ok := h.ensureUserID(operation)
	if !ok {
		return
	}

	h.processUploadOrder(operation, userID)
}

func (h *Handler) processUploadOrder(operation *api.OperationUploadOrder, userID domain.UserID) {
	orderID, ok := h.processUploadOrderRequest(operation)
	if !ok {
		return
	}

	err := h.orderService.CreateOrder(operation.Context(), userID, orderID)
	if err != nil {
		h.processUploadOrderError(operation, userID, err)

		return
	}

	operation.RespondAccepted()
}

func (h *Handler) processUploadOrderRequest(operation *api.OperationUploadOrder) (domain.OrderID, bool) {
	req, err := operation.GetRequest()
	if err != nil {
		h.processUploadOrderRequestError(operation, err)

		return domain.OrderID(0), false
	}

	return domain.OrderID(req), true
}

func (h *Handler) processUploadOrderRequestError(operation *api.OperationUploadOrder, err error) {
	if errors.Is(err, api.ErrOrderIDLuhnChecksumMismatch) {
		operation.RespondUnprocessableEntity()

		return
	}

	operation.RespondBadRequest()
}

func (h *Handler) processUploadOrderError(operation *api.OperationUploadOrder, userID domain.UserID, err error) {
	var errOrderExists *domain.OrderIDAlreadyExistsError

	if errors.As(err, &errOrderExists) {
		if errOrderExists.CreatedBy == userID {
			operation.RespondOK()
		} else {
			operation.RespondConflict()
		}

		return
	}

	operation.RespondServerError()
}
