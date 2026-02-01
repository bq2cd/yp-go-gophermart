package models

import (
	"time"
)

// Withdrawal is a representation of a withdrawal transaction made by a user.
type Withdrawal struct {
	ID          uint      `gorm:"primaryKey"`
	Amount      float64   `gorm:"not null;precision:16;scale:4"`
	CreatedAt   time.Time `gorm:"index:,sort:desc"`
	NextOrderID uint      `gorm:"uniqueIndex"`

	UserID uint
	User   User
}
