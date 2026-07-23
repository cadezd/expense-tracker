package transaction

import "errors"

var (
	ErrEmptyRequest             = errors.New("invalid empty request")
	ErrInvalidType              = errors.New("invalid transaction type")
	ErrInvalidAmount            = errors.New("invalid transaction amount")
	ErrNegativeAmount           = errors.New("negative transaction amount")
	ErrInvalidCurrency          = errors.New("invalid transaction currency")
	ErrInvalidTransactionDate   = errors.New("invalid transaction date")
	ErrReceiptOwnershipMismatch = errors.New("receipt belongs to a different user")

	ErrNotFound = errors.New("transaction not found")
)
