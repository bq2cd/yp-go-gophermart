package models

// Accrual is a representation of an order with accrual points.
type Accrual struct {
	Amount float64 `gorm:"not null;precision:16;scale:4"`

	OrderID uint `gorm:"uniqueIndex"`
}
