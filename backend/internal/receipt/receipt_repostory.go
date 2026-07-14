package receipt

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceiptRepository interface {
	Create(ctx context.Context, receipt *Receipt) (*Receipt, error)
	Update(ctx context.Context, receipt *Receipt) (*Receipt, error)
	GetByID(ctx context.Context, userID, receiptID uuid.UUID) (*Receipt, error)
	List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Receipt, error)
	Delete(ctx context.Context, userID, receiptID uuid.UUID) error
}
