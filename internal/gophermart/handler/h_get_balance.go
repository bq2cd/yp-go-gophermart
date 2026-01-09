package handler

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// GetBalance implements [api.OperationGetBalance].
func (h *Handler) GetBalance(operation *api.OperationGetBalance) {
	userID, ok := h.ensureUserID(operation)
	if !ok {
		return
	}

	balance, err := h.balanceService.GetBalance(userID)
	if err != nil {
		h.sendGetBalanceError(operation, err)

		return
	}

	withdrawn, err := h.balanceService.GetTotalAmountWithdrawn(userID)
	if err != nil {
		h.sendGetBalanceError(operation, err)

		return
	}

	operation.RespondOK(api.Balance{
		Current:   balance,
		Withdrawn: withdrawn,
	})
}

func (h *Handler) sendGetBalanceError(operation *api.OperationGetBalance, _ error) {
	operation.RespondServerError()
}
