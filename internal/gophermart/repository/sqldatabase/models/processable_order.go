package models

import (
	"time"
)

// ProcessableOrder is a representation of an order eligible for further processing.
// It contains information about retries, target processing date and links to the
// original [Order].
type ProcessableOrder struct {
	OrderID uint `gorm:"primaryKey;autoIncrement:false"`
	Order   Order

	ProcessAfter time.Time `gorm:"not null;index:,sort:asc"`
	Retries      uint      `gorm:"not null"`
}
