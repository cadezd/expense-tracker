package transaction

import (
	"context"

	"github.com/google/uuid"
)

type TransactionRepostory interface {
	Crate(ctx context.Context, transaction Transaction) (Transaction, error)
	Update(ctx context.Context, transaction Transaction) (Transaction, error)
	GetByID(ctx context.Context, userID, transactionID uuid.UUID) (Transaction, error)
	List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]Transaction, error)
	Delete(ctx context.Context, userID, transactionID uuid.UUID) error
}
