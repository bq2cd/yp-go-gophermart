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
	req, err := operation.GetRequest()
	if err != nil {
		operation.RespondBadRequest()

		return
	}

	orderID := domain.OrderID(req)

	err = h.orderService.CreateOrder(userID, orderID)
	if err != nil {
		h.processUploadOrderError(operation, userID, err)

		return
	}

	operation.RespondAccepted()
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

	if errors.Is(err, domain.ErrOrderIDValidationFailed) {
		operation.RespondUnprocessableEntity()

		return
	}

	operation.RespondServerError()
}
