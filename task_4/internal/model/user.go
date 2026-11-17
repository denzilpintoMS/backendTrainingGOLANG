package model

import "time"

// User represents an application user.
type User struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string    `json:"name" gorm:"not null;check:users_name_not_blank,length(trim(name)) > 0"`
	Email        string    `json:"email" gorm:"type:varchar(254);not null;uniqueIndex:idx_users_email,expression:lower(email);check:users_email_not_blank,length(trim(email)) > 0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
