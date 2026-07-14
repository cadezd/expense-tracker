package transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) (*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) (*Transaction, error)
	GetByID(ctx context.Context, userID, transactionID uuid.UUID) (*Transaction, error)
	List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Transaction, error)
	Delete(ctx context.Context, userID, transactionID uuid.UUID) error
}

type PostgresTransactionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresTransactionRepository(db *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{
		db: db,
	}
}

func (tr *PostgresTransactionRepository) Create(
	ctx context.Context,
	transaction *Transaction,
) (*Transaction, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (tr *PostgresTransactionRepository) Update(
	ctx context.Context,
	transaction *Transaction,
) (*Transaction, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (tr *PostgresTransactionRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	transactionID uuid.UUID,
) (*Transaction, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (tr *PostgresTransactionRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	offset int,
	limit int,
) ([]*Transaction, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (tr *PostgresTransactionRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
	transactionID uuid.UUID,
) error {
	return fmt.Errorf("Not implemented")
}
