package models

import (
	"time"
)

// Order is a representation of an order for a given user.
type Order struct {
	ID        uint      `gorm:"primaryKey;autoIncrement:false"`
	Status    int       `gorm:"not null;index:"`
	CreatedAt time.Time `gorm:"index:,sort:desc"`

	Accrual Accrual

	UserID uint
	User   User
}
