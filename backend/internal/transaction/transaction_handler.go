package transaction

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	service *TransactionService
}

func NewTransactionHandler(service *TransactionService) *TransactionHandler {
	return &TransactionHandler{
		service: service,
	}
}

func (th *TransactionHandler) Create(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_TRANSACTION_BODY", "request body is invalid"))
		return
	}

	transaction, err := th.service.Create(
		c.Request.Context(),
		userID,
		&req,
	)
	if err != nil {
		_ = c.Error(mapTransactionError(err, "CREATE_TRANSACTION_FAILED", "failed to create transaction"))
		return
	}

	common.Ok(c, transaction)
}

func (th *TransactionHandler) Update(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	var req UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_TRANSACTION_BODY", "request body is invalid"))
		return
	}

	transaction, err := th.service.Update(
		c.Request.Context(),
		userID,
		&req,
	)
	if err != nil {
		_ = c.Error(mapTransactionError(err, "UPDATE_TRANSACTION_FAILED", "failed to update transaction"))
		return
	}

	common.Ok(c, transaction)
}

func (th *TransactionHandler) GetByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_TRANSACTION_ID", "transaction id must be a valid UUID"))
		return
	}

	transaction, err := th.service.GetByID(
		c.Request.Context(),
		userID,
		transactionID,
	)
	if err != nil {
		_ = c.Error(mapTransactionError(err, "GET_TRANSACTION_FAILED", "failed to load transaction"))
		return
	}

	common.Ok(c, transaction)
}

func (th *TransactionHandler) List(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	offsetStr := c.Query("offset")
	if offsetStr == "" {
		_ = c.Error(common.NewBadRequestError("MISSING_OFFSET", "offset query parameter is required"))
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		_ = c.Error(common.NewBadRequestError("INVALID_OFFSET", "offset must be a non-negative integer"))
		return
	}

	limitStr := c.Query("limit")
	if limitStr == "" {
		_ = c.Error(common.NewBadRequestError("MISSING_LIMIT", "limit query parameter is required"))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		_ = c.Error(common.NewBadRequestError("INVALID_LIMIT", "limit must be an integer between 1 and 100"))
		return
	}

	transactions, err := th.service.List(
		c.Request.Context(),
		userID,
		offset,
		limit,
	)
	if err != nil {
		_ = c.Error(mapTransactionError(err, "LIST_TRANSACTIONS_FAILED", "failed to list transactions"))
		return
	}

	common.Ok(c, transactions)
}

func (th *TransactionHandler) Delete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_TRANSACTION_ID", "transaction id must be a valid UUID"))
		return
	}

	err = th.service.Delete(
		c.Request.Context(),
		userID,
		transactionID,
	)
	if err != nil {
		_ = c.Error(mapTransactionError(err, "DELETE_TRANSACTION_FAILED", "failed to delete transaction"))
		return
	}

	common.Ok(c, nil)
}

// -------------------
// HELPERS
// -------------------

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return userID, true
}

func mapTransactionError(err error, fallbackCode, fallbackMessage string) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return common.NewAppError(http.StatusNotFound, "TRANSACTION_NOT_FOUND", "transaction not found")
	case errors.Is(err, ErrReceiptOwnershipMismatch):
		return common.NewBadRequestError("INVALID_RECEIPT_REFERENCE", "receipt does not exist or is not accessible")
	case errors.Is(err, ErrEmptyRequest),
		errors.Is(err, ErrInvalidTransactionID),
		errors.Is(err, ErrInvalidObjectVersion),
		errors.Is(err, ErrNoFieldsToUpdate),
		errors.Is(err, ErrNullType),
		errors.Is(err, ErrInvalidType),
		errors.Is(err, ErrNullAmount),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrNegativeAmount),
		errors.Is(err, ErrNullCurrency),
		errors.Is(err, ErrInvalidCurrency),
		errors.Is(err, ErrNullTransactionDate),
		errors.Is(err, ErrInvalidTransactionDate),
		errors.Is(err, ErrNullCategory),
		errors.Is(err, ErrInvalidOffset),
		errors.Is(err, ErrInvalidLimit):
		return common.NewBadRequestError("INVALID_TRANSACTION_REQUEST", err.Error())
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return common.NewAppError(http.StatusRequestTimeout, "REQUEST_CANCELLED", "request was cancelled or timed out")
	default:
		return common.NewAppError(http.StatusInternalServerError, fallbackCode, fallbackMessage)
	}
}
