package models

// User owns one account. Password must contain a password hash.
type User struct {
	ID       uint     `json:"id"`
	Name     string   `json:"name"`
	Email    string   `gorm:"uniqueIndex" json:"email"`
	Password string   `json:"-"`
	Account  *Account `json:"account,omitempty"`
}
