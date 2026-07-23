package transaction

import "errors"

var (
	ErrEmptyRequest             = errors.New("invalid empty request")
	ErrInvalidTransactionID     = errors.New("invalid transaction id")
	ErrInvalidObjectVersion     = errors.New("invalid transaction object version")
	ErrNoFieldsToUpdate         = errors.New("no transaction fields to update")
	ErrNullType                 = errors.New("transaction type cannot be null")
	ErrInvalidType              = errors.New("invalid transaction type")
	ErrNullAmount               = errors.New("transaction amount cannot be null")
	ErrInvalidAmount            = errors.New("invalid transaction amount")
	ErrNegativeAmount           = errors.New("negative transaction amount")
	ErrNullCurrency             = errors.New("transaction currency cannot be null")
	ErrInvalidCurrency          = errors.New("invalid transaction currency")
	ErrNullTransactionDate      = errors.New("transaction date cannot be null")
	ErrInvalidTransactionDate   = errors.New("invalid transaction date")
	ErrNullCategory             = errors.New("transaction category cannot be null")
	ErrReceiptOwnershipMismatch = errors.New("receipt belongs to a different user")
	ErrUnknownUpdateField       = errors.New("unknown transaction update field")
	ErrInvalidOffset            = errors.New("invalid transaction offset")
	ErrInvalidLimit             = errors.New("invalid transaction limit")

	ErrNotFound = errors.New("transaction not found")
)
