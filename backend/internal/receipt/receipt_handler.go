package receipt

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/cadezd/expense-tracker/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReceiptHandler struct {
	service *ReceiptService
}

func NewReceiptHandler(service *ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		service: service,
	}
}

func (rh *ReceiptHandler) Upload(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(common.NewBadRequestError("FILE_REQUIRED", "multipart form field 'file' is required"))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_FILE", "uploaded file could not be opened"))
		return
	}
	defer file.Close()

	receipt, err := rh.service.Upload(
		c.Request.Context(),
		userID,
		fileHeader.Filename,
		file,
	)
	if err != nil {
		_ = c.Error(mapReceiptError(err, "UPLOAD_RECEIPT_FAILED", "failed to upload receipt"))
		return
	}

	common.Ok(c, receipt)
}

func (rh *ReceiptHandler) GetByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	receiptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_RECEIPT_ID", "receipt id must be a valid UUID"))
		return
	}

	receipt, err := rh.service.GetByID(
		c.Request.Context(),
		userID,
		receiptID,
	)
	if err != nil {
		_ = c.Error(mapReceiptError(err, "GET_RECEIPT_FAILED", "failed to load receipt"))
		return
	}

	common.Ok(c, receipt)
}

func (rh *ReceiptHandler) List(c *gin.Context) {
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

	receipts, err := rh.service.List(
		c.Request.Context(),
		userID,
		offset,
		limit,
	)
	if err != nil {
		_ = c.Error(mapReceiptError(err, "LIST_RECEIPTS_FAILED", "failed to list receipts"))
		return
	}

	common.Ok(c, receipts)
}

func (rh *ReceiptHandler) Delete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		_ = c.Error(common.ErrUnauthorized)
		return
	}

	receiptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_RECEIPT_ID", "receipt id must be a valid UUID"))
		return
	}

	err = rh.service.Delete(
		c.Request.Context(),
		userID,
		receiptID,
	)
	if err != nil {
		_ = c.Error(mapReceiptError(err, "DELETE_RECEIPT_FAILED", "failed to delete receipt"))
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

func mapReceiptError(err error, fallbackCode, fallbackMessage string) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return common.NewAppError(http.StatusNotFound, "RECEIPT_NOT_FOUND", "receipt not found")
	case errors.Is(err, storage.ErrFileTooLarge):
		return common.NewAppError(http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "uploaded file exceeds the allowed size")
	case errors.Is(err, storage.ErrUnsupportedMIMEType):
		return common.NewAppError(http.StatusUnsupportedMediaType, "UNSUPPORTED_MIME_TYPE", "unsupported file type")
	case errors.Is(err, storage.ErrEmptyFile):
		return common.NewAppError(http.StatusBadRequest, "EMPTY_FILE", "uploaded file is empty")
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return common.NewAppError(http.StatusRequestTimeout, "REQUEST_CANCELLED", "request was cancelled or timed out")
	default:
		return common.NewAppError(http.StatusInternalServerError, fallbackCode, fallbackMessage)
	}
}
