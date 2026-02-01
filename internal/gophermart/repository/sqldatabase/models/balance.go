package models

// Balance is a representation of user's balance and total withdrawals amount.
type Balance struct {
	ID        uint    `gorm:"primaryKey"`
	Current   float64 `gorm:"not null;precision:16;scale:4"`
	Withdrawn float64 `gorm:"not null;precision:16;scale:4"`

	UserID uint `gorm:"uniqueIndex"`
}
