package api

import (
	"fmt"
	"net/http"
)

// OperationUploadOrder is responsible for uploading a new order by specifying its [OrderID].
// Only available for authorized users.
type OperationUploadOrder struct {
	operationCtxSecure
}

// GetRequest decodes JSON body of the HTTP request into [OrderID]
// and performs necessary validation.
// An error is returned if decoding or validation fails.
func (op *OperationUploadOrder) GetRequest() (OrderID, error) {
	req, err := operationGetRequest[OrderID](op.ginCtx)
	if err != nil {
		return 0, fmt.Errorf("cannot parse order ID: %w", err)
	}

	err = req.Validate()
	if err != nil {
		return 0, fmt.Errorf("order ID validation error: %w", err)
	}

	return *req, nil
}

// RespondOK returns '200 OK' when an order with the same ID has been
// already uploaded by the user.
func (op *OperationUploadOrder) RespondOK() {
	op.ginCtx.Status(http.StatusOK)
}

// RespondAccepted returns '202 Accepted' when an order is registered for the further processing in the system.
func (op *OperationUploadOrder) RespondAccepted() {
	op.ginCtx.Status(http.StatusAccepted)
}

// RespondConflict returns '409 Conflict' when given order ID has been
// already uploaded by another user.
func (op *OperationUploadOrder) RespondConflict() {
	op.ginCtx.AbortWithStatus(http.StatusConflict)
}

// RespondUnprocessableEntity returns '422 Unprocessable Entity' when given order ID is incorrect, e.g.
// there is an error in one of the digits.
func (op *OperationUploadOrder) RespondUnprocessableEntity() {
	op.ginCtx.AbortWithStatus(http.StatusUnprocessableEntity)
}
