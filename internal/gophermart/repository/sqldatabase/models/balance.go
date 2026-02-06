package models

// Balance is a representation of user's balance and total withdrawals amount.
type Balance struct {
	UserID    uint    `gorm:"primaryKey;autoIncrement:false"`
	Current   float64 `gorm:"not null;precision:16;scale:4"`
	Withdrawn float64 `gorm:"not null;precision:16;scale:4"`
}
