package models

// User owns one account. Password must contain a password hash.
type User struct {
	ID       uint     `gorm:"primaryKey" json:"id"`
	Name     string   `gorm:"not null" json:"name"`
	Email    string   `gorm:"not null;uniqueIndex" json:"email"`
	Password string   `gorm:"not null" json:"-"`
	Account  *Account `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"account,omitempty"`
}
