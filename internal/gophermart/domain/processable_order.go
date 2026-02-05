package domain

// ProcessableOrder represents an order that needs further processing.
type ProcessableOrder struct {
	UserID  UserID
	OrderID OrderID
}
