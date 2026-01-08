package api

import (
	"net/http"
)

// OperationGetBalance is responsible for returning user's [Balance].
// Only available for authorized users.
type OperationGetBalance struct {
	operationCtxSecure
}

// RespondOK returns '200 OK' with JSON-encoded [Balance].
func (op *OperationGetBalance) RespondOK(resp Balance) {
	op.ginCtx.JSON(http.StatusOK, resp)
}
