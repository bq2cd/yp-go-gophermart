package handler

import (
	"errors"
	"strconv"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// Withdraw implements [api.OperationWithdraw].
func (h *Handler) Withdraw(operation *api.OperationWithdraw) {
	userID, ok := h.ensureUserID(operation)
	if !ok {
		return
	}

	h.processWithdraw(operation, userID)
}

func (h *Handler) processWithdraw(operation *api.OperationWithdraw, userID domain.UserID) {
	orderID, amount, ok := h.processWithdrawRequest(operation)
	if !ok {
		return
	}

	err := h.balanceService.PayForOrderFromBalance(userID, orderID, amount)
	if err != nil {
		h.processWithdrawError(operation, err)

		return
	}

	operation.RespondOK()
}

func (h *Handler) processWithdrawRequest(
	operation *api.OperationWithdraw,
) (domain.OrderID, float64, bool) {
	req, err := operation.GetRequest()
	if err != nil {
		operation.RespondUnprocessableEntity()

		return domain.OrderID(0), 0, false
	}

	orderNum, err := strconv.ParseUint(req.Order, 10, 64)
	if err != nil {
		operation.RespondUnprocessableEntity()

		return domain.OrderID(0), 0, false
	}

	return domain.OrderID(orderNum), req.Sum, true
}

func (h *Handler) processWithdrawError(operation *api.OperationWithdraw, err error) {
	if errors.Is(err, domain.ErrBalanceNotEnoughFunds) {
		operation.RespondPaymentRequired()

		return
	}

	operation.RespondServerError()
}
