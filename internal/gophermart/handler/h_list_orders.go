package handler

import (
	"strconv"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

// ListOrders implements [api.OperationListOrders].
func (h *Handler) ListOrders(operation *api.OperationListOrders) {
	userID, ok := h.ensureUserID(operation)
	if !ok {
		return
	}

	orders, err := h.orderService.GetOrders(userID)
	if err != nil {
		h.sendListOrdersError(operation, err)

		return
	}

	if len(orders) == 0 {
		operation.RespondNoContent()

		return
	}

	h.sendListOrdersOK(operation, userID, orders)
}

func (h *Handler) sendListOrdersError(operation *api.OperationListOrders, _ error) {
	operation.RespondServerError()
}

func (h *Handler) sendListOrdersOK(operation *api.OperationListOrders, userID domain.UserID, orders []domain.Order) {
	orderIDs := make([]domain.OrderID, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	accruals, err := h.orderService.GetAccruals(userID, orderIDs)
	if err != nil {
		h.sendListOrdersError(operation, err)

		return
	}

	apiOrders := make([]api.Order, 0, len(orders))
	for _, order := range orders {
		apiOrders = append(apiOrders, api.Order{
			Accrual:    accruals[order.ID],
			Number:     strconv.FormatUint(uint64(order.ID), 10),
			Status:     convertOrderStatus(order.Status),
			UploadedAt: order.CreatedAt,
		})
	}

	operation.RespondOK(apiOrders)
}

func convertOrderStatus(status domain.OrderStatus) api.OrderStatus {
	switch status {
	case domain.OrderStatusNew:
		return api.OrderStatusNew
	case domain.OrderStatusProcessing:
		return api.OrderStatusProcessing
	case domain.OrderStatusProcessed:
		return api.OrderStatusProcessed
	case domain.OrderStatusInvalid:
		return api.OrderStatusInvalid
	default:
		return api.OrderStatusInvalid
	}
}
