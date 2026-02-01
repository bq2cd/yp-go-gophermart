package models

// User is a representation of a user.
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Login        string `gorm:"not null;uniqueIndex"`
	PasswordHash []byte `gorm:"not null"`

	Balance     Balance
	Orders      []Order
	Withdrawals []Withdrawal
}
