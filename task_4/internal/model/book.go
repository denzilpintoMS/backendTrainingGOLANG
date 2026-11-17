package model

import "time"

type Book struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"not null;check:books_name_not_blank,length(trim(name)) > 0"`
	Author      string `json:"author" gorm:"not null;default:''"`
	Description string `json:"desc" gorm:"not null;default:''"`
	// Store audit timestamps without changing the existing book API response.
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
