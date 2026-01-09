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

	withdrawals, err := h.balanceService.GetWithdrawalTransactions(userID)
	if err != nil {
		h.sendListWithdrawalsError(operation)

		return
	}

	h.sendListWithdrawalsOK(operation, withdrawals)
}

func (h *Handler) sendListWithdrawalsError(operation *api.OperationListWithdrawals) {
	operation.RespondServerError()
}

func (h *Handler) sendListWithdrawalsOK(
	operation *api.OperationListWithdrawals,
	withdrawals []domain.WithdrawalTransaction,
) {
	if len(withdrawals) == 0 {
		operation.RespondNoContent()

		return
	}

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
