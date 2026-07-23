package transaction

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Type string

const (
	TypeExpense = "expense"
	TypeIncome  = "income"
)

// -------------------
// DOMAIN
// -------------------

type Transaction struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	ReceiptID       *uuid.UUID      `json:"receipt_id"`
	Type            Type            `json:"type"`
	Counterparty    *string         `json:"counterparty"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	TransactionDate time.Time       `json:"transaction_date"`
	Category        string          `json:"category"`
	Description     *string         `json:"description"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	ObjectVersion   int64           `json:"object_version"`
}

// -------------------
// DTOs
// -------------------

type CreateTransactionRequest struct {
	ReceiptID       *uuid.UUID `json:"receipt_id"`
	Type            Type       `json:"type" binding:"required"`
	Counterparty    *string    `json:"counterparty"`
	Amount          string     `json:"amount" binding:"required"`
	Currency        string     `json:"currency" binding:"required,len=3"`
	TransactionDate string     `json:"transaction_date" binding:"required"`
	Category        string     `json:"category"`
	Description     *string    `json:"description"`
}

type UpdateTransactionRequest struct {
	ReceiptID       *uuid.UUID `json:"receipt_id"`
	Type            *Type      `json:"type"`
	Counterparty    *string    `json:"counterparty"`
	Amount          *string    `json:"amount"`
	Currency        *string    `json:"currency"`
	TransactionDate *string    `json:"transaction_date"`
	Category        *string    `json:"category"`
	Description     *string    `json:"description"`
}
