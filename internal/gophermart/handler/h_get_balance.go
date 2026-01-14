package handler

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// GetBalance implements [api.OperationGetBalance].
func (h *Handler) GetBalance(operation *api.OperationGetBalance) {
	userID, ok := h.ensureUserID(operation)
	if !ok {
		return
	}

	h.processGetBalance(operation, userID)
}

func (h *Handler) processGetBalance(operation *api.OperationGetBalance, userID domain.UserID) {
	balance, err := h.balanceService.GetBalance(operation.Context(), userID)
	if err != nil {
		h.processGetBalanceError(operation, err)

		return
	}

	withdrawn, err := h.balanceService.GetTotalAmountWithdrawn(operation.Context(), userID)
	if err != nil {
		h.processGetBalanceError(operation, err)

		return
	}

	operation.RespondOK(api.Balance{
		Current:   balance,
		Withdrawn: withdrawn,
	})
}

func (h *Handler) processGetBalanceError(operation *api.OperationGetBalance, _ error) {
	operation.RespondServerError()
}
