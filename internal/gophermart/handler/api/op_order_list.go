package api

import (
	"net/http"
)

// OperationListOrders is responsible for returning an array of user's [Order] objects.
// Only available for authorized users.
type OperationListOrders struct {
	operationCtxSecure
}

// RespondOK returns '200 OK' with JSON-encoded array of [Order] items.
func (op *OperationListOrders) RespondOK(resp []Order) {
	op.ginCtx.JSON(http.StatusOK, resp)
}

// RespondNoContent returns '204 No Content' when a user has no orders.
func (op *OperationListOrders) RespondNoContent() {
	op.ginCtx.Status(http.StatusNoContent)
}
