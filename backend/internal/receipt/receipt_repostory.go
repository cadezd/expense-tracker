package receipt

import (
	"context"
	"fmt"

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

type PostgresReceiptRepository struct {
	db *pgxpool.Pool
}

func NewPostgresReceiptRepository(db *pgxpool.Pool) *PostgresReceiptRepository {
	return &PostgresReceiptRepository{
		db: db,
	}
}

func (rr *PostgresReceiptRepository) Create(
	ctx context.Context,
	receipt *Receipt,
) (*Receipt, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (rr *PostgresReceiptRepository) Update(
	ctx context.Context,
	receipt *Receipt,
) (*Receipt, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (rr *PostgresReceiptRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	receiptID uuid.UUID,
) (*Receipt, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (rr *PostgresReceiptRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	offset int,
	limit int,
) ([]*Receipt, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (rr *PostgresReceiptRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
	receiptID uuid.UUID,
) error {
	return fmt.Errorf("Not implemented")
}
