package models

// Accrual is a representation of an order with accrual points.
type Accrual struct {
	OrderID uint    `gorm:"primaryKey;autoIncrement:false"`
	Amount  float64 `gorm:"not null;precision:16;scale:4"`
}
