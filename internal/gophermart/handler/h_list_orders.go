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

	h.processListOrders(operation, userID)
}

func (h *Handler) processListOrders(operation *api.OperationListOrders, userID domain.UserID) {
	orders, err := h.orderService.GetOrders(userID)
	if err != nil {
		h.processListOrdersError(operation, err)

		return
	}

	if len(orders) == 0 {
		operation.RespondNoContent()

		return
	}

	h.createListOrdersResponse(operation, userID, orders)
}

func (h *Handler) processListOrdersError(operation *api.OperationListOrders, _ error) {
	operation.RespondServerError()
}

func (h *Handler) createListOrdersResponse(
	operation *api.OperationListOrders,
	userID domain.UserID,
	orders []domain.Order,
) {
	accruals, err := h.orderService.GetAccruals(userID, getOrderIDList(orders))
	if err != nil {
		h.processListOrdersError(operation, err)

		return
	}

	operation.RespondOK(convertOrdersToAPIResponse(orders, accruals))
}

func getOrderIDList(orders []domain.Order) []domain.OrderID {
	orderIDs := make([]domain.OrderID, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	return orderIDs
}

func convertOrdersToAPIResponse(orders []domain.Order, accruals map[domain.OrderID]float64) []api.Order {
	apiOrders := make([]api.Order, 0, len(orders))

	for _, order := range orders {
		apiOrders = append(apiOrders, api.Order{
			Accrual:    accruals[order.ID],
			Number:     strconv.FormatUint(uint64(order.ID), 10),
			Status:     convertOrderStatus(order.Status),
			UploadedAt: order.CreatedAt,
		})
	}

	return apiOrders
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
