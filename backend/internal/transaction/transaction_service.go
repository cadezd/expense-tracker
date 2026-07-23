package transaction

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/cadezd/expense-tracker/internal/receipt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionService struct {
	transactionRepository TransactionRepository
	receiptRepository     receipt.ReceiptRepository
}

func NewTransactionService(
	transactionRepository TransactionRepository,
	receiptRepository receipt.ReceiptRepository,
) *TransactionService {
	return &TransactionService{
		transactionRepository: transactionRepository,
		receiptRepository:     receiptRepository,
	}
}

func (ts *TransactionService) Create(
	ctx context.Context,
	userID uuid.UUID,
	req *CreateTransactionRequest,
) (*Transaction, error) {
	if req == nil {
		return nil, ErrEmptyRequest
	}

	if req.ReceiptID != nil && *req.ReceiptID != uuid.Nil {
		_, err := ts.receiptRepository.GetByID(ctx, userID, *req.ReceiptID)
		if errors.Is(err, receipt.ErrNotFound) {
			return nil, ErrReceiptOwnershipMismatch
		}
		if err != nil {
			return nil, fmt.Errorf("load receipt for transaction: %w", err)
		}
	}

	if !slices.Contains([]Type{TypeExpense, TypeIncome}, req.Type) {
		return nil, ErrInvalidType
	}

	if req.Amount.LessThan(decimal.NewFromInt(0)) {
		return nil, ErrNegativeAmount
	}

	if len(req.Currency) != 3 {
		return nil, ErrInvalidCurrency
	}

	transaction := &Transaction{
		UserID:          userID,
		ReceiptID:       req.ReceiptID,
		Type:            req.Type,
		Counterparty:    req.Counterparty,
		Amount:          req.Amount,
		Currency:        req.Currency,
		TransactionDate: req.TransactionDate,
		Category:        req.Category,
		Description:     req.Description,
	}
	err := ts.transactionRepository.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("create transaction record: %w", err)
	}

	return transaction, nil
}

func (ts *TransactionService) Update(
	ctx context.Context,
	userID uuid.UUID,
	req *UpdateTransactionRequest,
) (*Transaction, error) {
	if req == nil {
		return nil, ErrEmptyRequest
	}

	if req.ID == uuid.Nil {
		return nil, ErrInvalidTransactionID
	}

	if req.ObjectVersion < 1 {
		return nil, ErrInvalidObjectVersion
	}

	hasUpdates := false

	if req.ReceiptID.IsSet() && !req.ReceiptID.IsNull() {
		hasUpdates = true
		receiptID, _ := req.ReceiptID.Value()

		_, err := ts.receiptRepository.GetByID(ctx, userID, receiptID)
		if errors.Is(err, receipt.ErrNotFound) {
			return nil, ErrReceiptOwnershipMismatch
		}
		if err != nil {
			return nil, fmt.Errorf("load receipt for transaction: %w", err)
		}
	}

	if req.Type.IsSet() {
		hasUpdates = true
		if req.Type.IsNull() {
			return nil, ErrNullType
		}

		transactionType, _ := req.Type.Value()
		if !slices.Contains([]Type{TypeExpense, TypeIncome}, transactionType) {
			return nil, ErrInvalidType
		}
	}

	if req.Amount.IsSet() {
		hasUpdates = true
		if req.Amount.IsNull() {
			return nil, ErrNullAmount
		}

		amount, _ := req.Amount.Value()
		if amount.LessThan(decimal.NewFromInt(0)) {
			return nil, ErrNegativeAmount
		}
	}

	if req.Currency.IsSet() {
		hasUpdates = true
		if req.Currency.IsNull() {
			return nil, ErrNullCurrency
		}

		curreny, _ := req.Currency.Value()
		if len(curreny) != 3 {
			return nil, ErrInvalidCurrency
		}
	}

	if req.Category.IsSet() {
		hasUpdates = true
		if req.Category.IsNull() {
			return nil, ErrNullCategory
		}
	}

	if req.TransactionDate.IsSet() {
		hasUpdates = true
		if req.TransactionDate.IsNull() {
			return nil, ErrNullTransactionDate
		}
	}

	if req.Counterparty.IsSet() {
		hasUpdates = true
	}

	if req.Description.IsSet() {
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, ErrNoFieldsToUpdate
	}

	transaction, err := ts.transactionRepository.GetByID(
		ctx,
		userID,
		req.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("load transaction for update: %w", err)
	}

	if req.ReceiptID.IsSet() {
		if req.ReceiptID.IsNull() {
			transaction.ReceiptID = nil
		} else {
			receiptID, _ := req.ReceiptID.Value()
			transaction.ReceiptID = &receiptID
		}
	}

	if req.Type.IsSet() {
		transactionType, _ := req.Type.Value()
		transaction.Type = transactionType
	}

	if req.Counterparty.IsSet() {
		if req.Counterparty.IsNull() {
			transaction.Counterparty = nil
		} else {
			counterparty, _ := req.Counterparty.Value()
			transaction.Counterparty = &counterparty
		}
	}

	if req.Amount.IsSet() {
		amount, _ := req.Amount.Value()
		transaction.Amount = amount
	}

	if req.Currency.IsSet() {
		curreny, _ := req.Currency.Value()
		transaction.Currency = curreny
	}

	if req.TransactionDate.IsSet() {
		transactionDate, _ := req.TransactionDate.Value()
		transaction.TransactionDate = transactionDate
	}

	if req.Category.IsSet() {
		category, _ := req.Category.Value()
		transaction.Category = category
	}

	if req.Description.IsSet() {
		if req.Description.IsNull() {
			transaction.Description = nil
		} else {
			description, _ := req.Description.Value()
			transaction.Description = &description
		}
	}

	transaction.ObjectVersion = req.ObjectVersion

	err = ts.transactionRepository.Update(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("update transaction record: %w", err)
	}

	return transaction, nil
}

func (ts *TransactionService) List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Transaction, error) {
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	if limit < 1 || limit > 100 {
		return nil, ErrInvalidLimit
	}

	transactions, err := ts.transactionRepository.List(
		ctx,
		userID,
		offset,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	return transactions, nil
}

func (ts *TransactionService) GetByID(ctx context.Context, userID, transactionID uuid.UUID) (*Transaction, error) {
	transaction, err := ts.transactionRepository.GetByID(
		ctx,
		userID,
		transactionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	return transaction, nil
}

func (ts *TransactionService) Delete(ctx context.Context, userID, transactionID uuid.UUID) error {
	err := ts.transactionRepository.Delete(
		ctx,
		userID,
		transactionID,
	)
	if err != nil {
		return fmt.Errorf("delete transaction record: %w", err)
	}

	return nil

}
