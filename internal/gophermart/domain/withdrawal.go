package domain

import "time"

// WithdrawalTransaction describes an action of using funds on user's balance
// to pay for a new order.
type WithdrawalTransaction struct {
	OrderID     OrderID
	Amount      float64
	ProcessedAt time.Time
}
