package api

import (
	"net/http"
)

// OperationListWithdrawals is responsible for returning an array of user's [WithdrawalTransaction] objects.
// Only available for authorized users.
type OperationListWithdrawals struct {
	operationCtxSecure
}

// RespondOK returns '200 OK' with JSON-encoded array of [WithdrawalTransaction].
func (op *OperationListWithdrawals) RespondOK(resp []WithdrawalTransaction) {
	op.ginCtx.JSON(http.StatusOK, resp)
}

// RespondNoContent returns '204 No Content' when user has not performed any withdrawals.
func (op *OperationListWithdrawals) RespondNoContent() {
	op.ginCtx.Status(http.StatusNoContent)
}
