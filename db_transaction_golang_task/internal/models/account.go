package models

import (
	"time"

	"gorm.io/gorm"
)

// Account belongs to a user. The unique user ID allows only one account per user.
type Account struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	AccountNumber string         `gorm:"not null;uniqueIndex" json:"account_number"`
	Balance       int64          `gorm:"not null;default:0;check:balance >= 0" json:"balance"` // Amount in paise.
	Status        string         `gorm:"not null;default:active;check:status IN ('active','frozen','closed')" json:"status"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
