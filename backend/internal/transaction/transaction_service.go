package transaction

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cadezd/expense-tracker/internal/receipt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionService struct {
	transactionRepository TransactionRepository
	receiptRepository     receipt.ReceiptRepository
}

func NewTransactionService(transactionRepository TransactionRepository, receiptRepository receipt.ReceiptRepository) *TransactionService {
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

	normalizedType := Type(strings.TrimSpace(strings.ToLower(string(req.Type))))
	if !slices.Contains([]Type{TypeExpense, TypeIncome}, normalizedType) {
		return nil, ErrInvalidType
	}

	amount, err := decimal.NewFromString(strings.TrimSpace(req.Amount))
	if err != nil {
		return nil, ErrInvalidAmount
	}

	if amount.LessThan(decimal.NewFromInt(0)) {
		return nil, ErrNegativeAmount
	}

	normalizedCurrency := strings.TrimSpace(strings.ToUpper(req.Currency))
	if len(normalizedCurrency) != 3 {
		return nil, ErrInvalidCurrency
	}

	transactionDate, err := time.Parse(time.RFC3339, strings.TrimSpace(req.TransactionDate))
	if err != nil {
		return nil, ErrInvalidTransactionDate
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

	transaction := &Transaction{
		UserID:          userID,
		ReceiptID:       req.ReceiptID,
		Type:            normalizedType,
		Counterparty:    req.Counterparty,
		Amount:          amount,
		Currency:        normalizedCurrency,
		TransactionDate: transactionDate,
		Category:        req.Category,
		Description:     req.Description,
	}
	err = ts.transactionRepository.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("create transaction record: %w", err)
	}

	return transaction, nil
}

func (ts *TransactionService) Update(ctx context.Context, userID uuid.UUID, req *UpdateTransactionRequest) {

}

func (ts *TransactionService) List(ctx context.Context, userID uuid.UUID, offset, limit int) {

}

func (ts *TransactionService) GetByID(ctx context.Context, userID, transactionID uuid.UUID) {

}

func (ts *TransactionService) Delete(ctx context.Context, userID, transactionID uuid.UUID) {

}
