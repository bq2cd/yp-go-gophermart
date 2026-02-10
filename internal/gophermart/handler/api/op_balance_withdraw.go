package api

import (
	"net/http"
)

// OperationWithdraw is responsible for creating a [WithdrawalRequest].
// Only available for authorized users.
type OperationWithdraw struct {
	operationCtxSecure
}

// GetRequest decodes JSON body of the HTTP request into [WithdrawalRequest]
// and performs necessary validation.
// An error is returned if decoding or validation fails.
func (op *OperationWithdraw) GetRequest() (*WithdrawalRequest, error) {
	return operationGetRequest[WithdrawalRequest](op.ginCtx)
}

// RespondOK returns '200 OK' when [WithdrawalRequest] is successful.
func (op *OperationWithdraw) RespondOK() {
	op.ginCtx.Status(http.StatusOK)
}

// RespondPaymentRequired returns '402 Payment Required' when [WithdrawalRequest.Sum] exceeds [Balance.Current].
func (op *OperationWithdraw) RespondPaymentRequired() {
	op.ginCtx.AbortWithStatus(http.StatusPaymentRequired)
}

// RespondUnprocessableEntity returns '422 Unprocessable Entity' when [WithdrawalRequest] contains invalid data,
// such as negative amount to withdraw or incorrect order id.
func (op *OperationWithdraw) RespondUnprocessableEntity() {
	op.ginCtx.AbortWithStatus(http.StatusUnprocessableEntity)
}
