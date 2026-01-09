package api

import (
	"time"
)

// Balance represents current value of the user's balance
// and the amount of previously withdrawn funds.
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// WithdrawalRequest is sent to the server to withdraw certain amount
// from the [Balance.Current] and pay for the given order ID.
type WithdrawalRequest struct {
	Order string  `json:"order" validate:"required,number,orderID"`
	Sum   float64 `json:"sum"   validate:"required,gt=0"`
}

// WithdrawalTransaction represents a successfully completed [WithdrawalRequest].
type WithdrawalTransaction struct {
	Order       string    `json:"order"`
	ProcessedAt time.Time `json:"processed_at"` //nolint:tagliatelle
	Sum         float64   `json:"sum"`
}
