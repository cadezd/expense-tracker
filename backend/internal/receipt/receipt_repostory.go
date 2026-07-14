package receipt

import (
	"context"

	"github.com/google/uuid"
)

type ReceiptRepository interface {
	Crate(ctx context.Context, receipt Receipt) (Receipt, error)
	Update(ctx context.Context, receipt Receipt) (Receipt, error)
	GetByID(ctx context.Context, userID, receiptID uuid.UUID) (Receipt, error)
	List(ctx context.Context, userID string, offset, limit int) ([]Receipt, error)
	Delete(ctx context.Context, userID, receiptID uuid.UUID) error
}
