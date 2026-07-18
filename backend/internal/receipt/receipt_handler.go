package receipt

import (
	"strconv"

	"github.com/cadezd/expense-tracker/internal/common"
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
		_ = c.Error(common.NewBadRequestError("FILE_REQUIRED", "File is required"))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_FILE", "Could not open uploaded file"))
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
		_ = c.Error(err)
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
		_ = c.Error(common.NewBadRequestError("MISSING_ID", "receipt_id filed is required"))
		return
	}

	receipt, err := rh.service.GetByID(
		c.Request.Context(),
		userID,
		receiptID,
	)
	if err != nil {
		_ = c.Error(err)
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
		_ = c.Error(common.NewBadRequestError("MISSING_PARAMETER", "offset is required"))
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_VALUE", "offset must be of type int"))
		return
	}

	limitStr := c.Query("limit")
	if limitStr == "" {
		_ = c.Error(common.NewBadRequestError("MISSING_PARAMETER", "limit is required"))
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		_ = c.Error(common.NewBadRequestError("INVALID_VALUE", "limit must be of type int"))
		return
	}

	receipts, err := rh.service.List(
		c.Request.Context(),
		userID,
		offset,
		limit,
	)
	if err != nil {
		_ = c.Error(err)
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
		_ = c.Error(common.NewBadRequestError("MISSING_ID", "receipt_id filed is required"))
		return
	}

	err = rh.service.Delete(
		c.Request.Context(),
		userID,
		receiptID,
	)
	if err != nil {
		_ = c.Error(err)
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
