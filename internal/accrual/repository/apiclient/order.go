package apiclient

import (
	"fmt"
	"strconv"

	"github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
)

// OrderStatus represents the status of an order as returned by the accrual system.
type OrderStatus string

const (
	// OrderStatusInvalid marks order as not valid for any processing by the accrual system.
	// This is a final status.
	OrderStatusInvalid OrderStatus = "INVALID"
	// OrderStatusRegistered marks order as registered for further processing by the accrual system.
	OrderStatusRegistered OrderStatus = "REGISTERED"
	// OrderStatusProcessing marks order as being processed by the accrual system.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusProcessed marks order as fully processed by the accrual system.
	// This is a final status.
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

// OrderResponse represents a successful response from the accrual system
// when sending a GET request to /api/orders/{order_id} endpoint.
type OrderResponse struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual float64     `json:"accrual"`
}

// ToOrder performs a conversion from [OrderResponse] to [domain.Order].
// It will return an error if order ID cannot be converted to a number,
// or if order status is unknown.
func (resp OrderResponse) ToOrder() (domain.Order, error) {
	var order domain.Order

	orderNumber, err := strconv.ParseUint(resp.Order, 10, 64)
	if err != nil {
		return order, fmt.Errorf("cannot convert order number to uint64: %w", err)
	}

	order.ID = domain.OrderID(orderNumber)

	switch resp.Status {
	case OrderStatusInvalid:
		order.Status = domain.OrderStatusInvalid
	case OrderStatusRegistered:
		order.Status = domain.OrderStatusRegistered
	case OrderStatusProcessing:
		order.Status = domain.OrderStatusProcessing
	case OrderStatusProcessed:
		order.Status = domain.OrderStatusProcessed
	default:
		return order, ErrUnknownOrderStatus
	}

	order.AccrualPoints = resp.Accrual

	return order, nil
}

// ConvertOrderToOrderResponse converts [domain.Order] to [OrderResponse].
// This function is primarily used in tests.
func ConvertOrderToOrderResponse(order domain.Order) (OrderResponse, error) {
	var resp OrderResponse

	resp.Order = order.ID.String()
	resp.Accrual = order.AccrualPoints

	switch order.Status {
	case domain.OrderStatusInvalid:
		resp.Status = OrderStatusInvalid
	case domain.OrderStatusRegistered:
		resp.Status = OrderStatusRegistered
	case domain.OrderStatusProcessing:
		resp.Status = OrderStatusProcessing
	case domain.OrderStatusProcessed:
		resp.Status = OrderStatusProcessed
	default:
		return resp, ErrUnknownOrderStatus
	}

	return resp, nil
}
