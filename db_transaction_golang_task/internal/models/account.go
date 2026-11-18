package models

import (
	"time"

	"gorm.io/gorm"
)

// Account belongs to a user. The unique user ID allows only one account per user.
type Account struct {
	ID            uint           `json:"id"`
	UserID        uint           `gorm:"not null;uniqueIndex" json:"user_id"`
	AccountNumber string         `gorm:"uniqueIndex" json:"account_number"`
	Balance       int64          `gorm:"check:balance >= 0" json:"balance"` // Amount in paise.
	Status        string         `gorm:"default:active" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at"`
}
