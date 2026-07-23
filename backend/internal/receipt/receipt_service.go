package receipt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

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
	storedFileMeta, err := rs.storage.Save(
		ctx,
		originalFilename,
		reader,
	)
	if err != nil {
		return nil, fmt.Errorf("save receipt file: %w", err)
	}

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
		cleanupErr := rs.storage.Delete(ctx, storedFileMeta.RelativePath)
		if cleanupErr != nil {
			return nil, errors.Join(
				fmt.Errorf("create receipt record: %w", err),
				fmt.Errorf("cleanup stored receipt file %q: %w", storedFileMeta.RelativePath, cleanupErr),
			)
		}

		return nil, fmt.Errorf("create receipt record: %w", err)
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
		return nil, fmt.Errorf("get receipt by id: %w", err)
	}

	return receipt, nil
}

func (rs *ReceiptService) GetFileByID(
	ctx context.Context,
	userID uuid.UUID,
	receiptID uuid.UUID,
) (*Receipt, io.ReadCloser, error) {
	receipt, err := rs.repository.GetByID(ctx, userID, receiptID)
	if err != nil {
		return nil, nil, fmt.Errorf("get receipt file metadata: %w", err)
	}

	file, err := rs.storage.Open(ctx, receipt.StoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("open receipt file %q: %w", receipt.StoragePath, err)
		}

		return nil, nil, fmt.Errorf("open receipt file %q: %w", receipt.StoragePath, err)
	}

	return receipt, file, nil
}

func (rs *ReceiptService) List(
	ctx context.Context,
	userID uuid.UUID,
	offset int,
	limit int,
) ([]*Receipt, error) {
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	if limit < 1 || limit > 100 {
		return nil, ErrInvalidLimit
	}

	receipts, err := rs.repository.List(
		ctx,
		userID,
		offset,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list receipts: %w", err)
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
		return fmt.Errorf("load receipt for delete: %w", err)
	}

	err = rs.storage.Delete(
		ctx,
		receipt.StoragePath,
	)
	if err != nil {
		return fmt.Errorf("delete stored receipt file %q: %w", receipt.StoragePath, err)
	}

	err = rs.repository.Delete(
		ctx,
		userID,
		receiptID,
	)
	if err != nil {
		return fmt.Errorf("delete receipt record: %w", err)
	}

	return nil
}
