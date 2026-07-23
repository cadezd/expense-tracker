package transaction

import (
	"time"

	"github.com/cadezd/expense-tracker/internal/common"
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
	ReceiptID       *uuid.UUID      `json:"receipt_id"`
	Type            Type            `json:"type"`
	Counterparty    *string         `json:"counterparty"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	TransactionDate time.Time       `json:"transaction_date"`
	Category        string          `json:"category"`
	Description     *string         `json:"description"`
}

type UpdateTransactionRequest struct {
	ID              uuid.UUID                        `json:"id"`
	ReceiptID       common.Optional[uuid.UUID]       `json:"receipt_id"`
	Type            common.Optional[Type]            `json:"type"`
	Counterparty    common.Optional[string]          `json:"counterparty"`
	Amount          common.Optional[decimal.Decimal] `json:"amount"`
	Currency        common.Optional[string]          `json:"currency"`
	TransactionDate common.Optional[time.Time]       `json:"transaction_date"`
	Category        common.Optional[string]          `json:"category"`
	Description     common.Optional[string]          `json:"description"`
	ObjectVersion   int64                            `json:"object_version"`
}
