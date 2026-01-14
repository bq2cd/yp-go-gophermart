package handler

import (
	"strconv"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// ListWithdrawals implements [api.OperationListWithdrawals].
func (h *Handler) ListWithdrawals(operation *api.OperationListWithdrawals) {
	userID, ok := h.ensureUserID(operation)
	if !ok {
		return
	}

	h.processListWithdrawals(operation, userID)
}

func (h *Handler) processListWithdrawals(operation *api.OperationListWithdrawals, userID domain.UserID) {
	withdrawals, err := h.balanceService.GetWithdrawalTransactions(operation.Context(), userID)
	if err != nil {
		h.processListWithdrawalsError(operation, err)

		return
	}

	if len(withdrawals) == 0 {
		operation.RespondNoContent()

		return
	}

	h.createListWithdrawalsResponse(operation, withdrawals)
}

func (h *Handler) processListWithdrawalsError(operation *api.OperationListWithdrawals, _ error) {
	operation.RespondServerError()
}

func (h *Handler) createListWithdrawalsResponse(
	operation *api.OperationListWithdrawals,
	withdrawals []domain.WithdrawalTransaction,
) {
	apiWithdrawals := make([]api.WithdrawalTransaction, 0, len(withdrawals))

	for _, item := range withdrawals {
		apiWithdrawals = append(apiWithdrawals, api.WithdrawalTransaction{
			Order:       strconv.FormatUint(uint64(item.OrderID), 10),
			ProcessedAt: item.ProcessedAt,
			Sum:         item.Amount,
		})
	}

	operation.RespondOK(apiWithdrawals)
}
