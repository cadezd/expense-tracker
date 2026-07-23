package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	Update(ctx context.Context, transaction *Transaction) error
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
) error {
	sqlQuery := `
		INSERT INTO transactions (user_id, receipt_id, type, counterparty, amount, currency, transaction_date, category, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at, object_version;
	`

	err := tr.db.QueryRow(ctx, sqlQuery,
		transaction.UserID,
		transaction.ReceiptID,
		transaction.Type,
		transaction.Counterparty,
		transaction.Amount,
		transaction.Currency,
		transaction.TransactionDate,
		transaction.Category,
		transaction.Description,
	).Scan(
		&transaction.ID,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
		&transaction.ObjectVersion,
	)
	if err != nil {
		return fmt.Errorf("transaction create error: %w", err)
	}

	return nil
}

func (tr *PostgresTransactionRepository) Update(
	ctx context.Context,
	transaction *Transaction,
) error {
	sqlQuery := `
		UPDATE transactions
		SET
			receipt_id = $1,
			type = $2,
			counterparty = $3,
			amount = $4,
			currency = $5,
			transaction_date = $6,
			category = $7,
			description = $8,
			updated_at = NOW(),
			object_version = object_version + 1
		WHERE
			user_id = $9 AND
			id = $10 AND
			object_version = $11
		RETURNING
			updated_at,
			object_version;
	`

	err := tr.db.QueryRow(ctx, sqlQuery,
		transaction.ReceiptID,
		transaction.Type,
		transaction.Counterparty,
		transaction.Amount,
		transaction.Currency,
		transaction.TransactionDate,
		transaction.Category,
		transaction.Description,
		transaction.UserID,
		transaction.ID,
		transaction.ObjectVersion,
	).Scan(
		&transaction.UpdatedAt,
		&transaction.ObjectVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update transaction error: %w", err)
	}

	return nil
}

func (tr *PostgresTransactionRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	transactionID uuid.UUID,
) (*Transaction, error) {
	transaction := &Transaction{}

	sqlQuery := `
		SELECT
			id,
			user_id,
			receipt_id,
			type,
			counterparty,
			amount,
			currency,
			transaction_date,
			category,
			description,
			created_at,
			updated_at,
			object_version
		FROM 
			transactions
		WHERE
			user_id = $1 AND
			id = $2;
	`

	err := tr.db.QueryRow(ctx, sqlQuery,
		userID,
		transactionID,
	).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.ReceiptID,
		&transaction.Type,
		&transaction.Counterparty,
		&transaction.Amount,
		&transaction.Currency,
		&transaction.TransactionDate,
		&transaction.Category,
		&transaction.Description,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
		&transaction.ObjectVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction by id error: %w", err)
	}

	return transaction, nil
}

func (tr *PostgresTransactionRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	offset int,
	limit int,
) ([]*Transaction, error) {
	sqlQuery := `
		SELECT
			id,
			user_id,
			receipt_id,
			type,
			counterparty,
			amount,
			currency,
			transaction_date,
			category,
			description,
			created_at,
			updated_at,
			object_version
		FROM 
			transactions
		WHERE
			user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3;
	`

	rows, err := tr.db.Query(ctx, sqlQuery,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list transactions error: %w", err)
	}
	defer rows.Close()

	transactions := make([]*Transaction, 0)
	for rows.Next() {

		transaction := &Transaction{}
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.ReceiptID,
			&transaction.Type,
			&transaction.Counterparty,
			&transaction.Amount,
			&transaction.Currency,
			&transaction.TransactionDate,
			&transaction.Category,
			&transaction.Description,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.ObjectVersion,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transaction error: %w", err)
		}

		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions error: %w", err)
	}

	return transactions, nil
}

func (tr *PostgresTransactionRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
	transactionID uuid.UUID,
) error {
	sqlQuery := `
		DELETE 
		FROM transactions
		WHERE 
			user_id = $1 AND
			id = $2;
	`

	_, err := tr.db.Exec(ctx, sqlQuery,
		userID,
		transactionID,
	)
	if err != nil {
		return fmt.Errorf("delete transation error: %w", err)
	}

	return nil
}
