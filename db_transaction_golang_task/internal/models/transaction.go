package models

import "time"

// Transaction records a deposit, withdrawal, or transfer.
type Transaction struct {
	ID            uint      `json:"id"`
	Type          string    `json:"type"`   // deposit, withdrawal, or transfer
	Amount        int64     `json:"amount"` // Amount in paise.
	FromAccountID *uint     `json:"from_account_id"`
	ToAccountID   *uint     `json:"to_account_id"`
	FromAccount   *Account  `json:"from_account,omitempty"`
	ToAccount     *Account  `json:"to_account,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
