package receipt

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/cadezd/expense-tracker/internal/storage"
	"github.com/google/uuid"
)

type ReceiptService struct {
	repository ReceiptRepository
	storage    storage.Storage
}

func NewReceiptService(repository ReceiptRepository, storage storage.Storage) *ReceiptService {
	return &ReceiptService{
		repository: repository,
		storage:    storage,
	}
}

func (rs *ReceiptService) Upload(
	ctx context.Context,
	userID uuid.UUID,
	originalFilename string,
	reader io.Reader,
) (*Receipt, error) {
	// Save to disk
	storedFileMeta, err := rs.storage.Save(
		ctx,
		originalFilename,
		reader,
	)
	if err != nil {
		return nil, mapToAppError(err)
	}

	// Save to db
	receipt := &Receipt{
		UserID:           userID,
		OriginalFilename: originalFilename,
		StoredFilename:   storedFileMeta.StoredFilename,
		StoragePath:      storedFileMeta.RelativePath,
		MimeType:         storedFileMeta.MIMEType,
		FileSize:         &storedFileMeta.Size,
	}

	err = rs.repository.Create(
		ctx,
		receipt,
	)
	if err != nil {
		_ = rs.storage.Delete(ctx, storedFileMeta.RelativePath)
		return nil, mapToAppError(err)
	}

	return receipt, nil
}

func (rs *ReceiptService) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	receiptID uuid.UUID,
) (*Receipt, error) {
	receipt, err := rs.repository.GetByID(
		ctx,
		userID,
		receiptID,
	)
	if err != nil {
		return nil, mapToAppError(err)
	}

	return receipt, nil
}

func (rs *ReceiptService) List(
	ctx context.Context,
	userID uuid.UUID,
	offset int,
	limit int,
) ([]*Receipt, error) {
	if offset < 0 {
		return nil, common.NewBadRequestError("INVALID_VALUE", "offset must be greather than 0")
	}

	if limit < 0 {
		return nil, common.NewBadRequestError("INVALID_VALUE", "limit must be netween 0 and 100")
	}

	if limit > 100 {
		return nil, common.NewBadRequestError("INVALID_VALUE", "limit must be netween 0 and 100")
	}

	receipts, err := rs.repository.List(
		ctx,
		userID,
		offset,
		limit,
	)
	if err != nil {
		return nil, mapToAppError(err)
	}

	return receipts, nil
}

func (rs *ReceiptService) Delete(
	ctx context.Context,
	userID uuid.UUID,
	receiptID uuid.UUID,
) error {
	receipt, err := rs.repository.GetByID(
		ctx,
		userID,
		receiptID,
	)
	if err != nil {
		return mapToAppError(err)
	}

	err = rs.storage.Delete(
		ctx,
		receipt.StoragePath,
	)
	if err != nil {
		return mapToAppError(err)
	}

	err = rs.repository.Delete(
		ctx,
		userID,
		receiptID,
	)
	if err != nil {
		return mapToAppError(err)
	}

	return nil
}

// -------------------
// HELPERS
// -------------------

func mapToAppError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return common.ErrNotFound
	case errors.Is(err, storage.ErrFileTooLarge):
		return common.NewAppError(http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "file is too large")
	case errors.Is(err, storage.ErrUnsupportedMIMEType):
		return common.NewAppError(http.StatusUnsupportedMediaType, "UNSUPPORTED_MIME_TYPE", "unsupported file type")
	case errors.Is(err, storage.ErrEmptyFile):
		return common.NewBadRequestError("EMPTY_FILE", "file is empty")
	case errors.Is(err, storage.ErrInvalidPath):
		return common.NewBadRequestError("INVALID_PATH", "invalid file path")
	default:
		return common.InternalError
	}
}
