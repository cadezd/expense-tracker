package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/cadezd/expense-tracker/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func resetDB(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(ctx, "TRUNCATE TABLE transactions, receipts, users RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func seedUser(t *testing.T, ctx context.Context, db *pgxpool.Pool, user *user.User) {
	t.Helper()

	err := db.QueryRow(ctx,
		`INSERT INTO users (email)
		VALUES ($1)
		RETURNING id, created_at, updated_at, object_version;`,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.ObjectVersion,
	)
	require.NoError(t, err)
}

func insertTransaction(t *testing.T, ctx context.Context, db *pgxpool.Pool, transaction *Transaction) {
	t.Helper()

	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = time.Now().UTC()
	}
	if transaction.UpdatedAt.IsZero() {
		transaction.UpdatedAt = transaction.CreatedAt
	}

	err := db.QueryRow(ctx,
		`INSERT INTO transactions (
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
		)
		VALUES (
			COALESCE($1, gen_random_uuid()),
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13
		)
		RETURNING id;`,
		nullableUUIDValue(transaction.ID),
		transaction.UserID,
		nullableUUID(transaction.ReceiptID),
		transaction.Type,
		nullableString(transaction.Counterparty),
		transaction.Amount,
		transaction.Currency,
		transaction.TransactionDate,
		transaction.Category,
		nullableString(transaction.Description),
		transaction.CreatedAt,
		transaction.UpdatedAt,
		int64(1),
	).Scan(&transaction.ID)
	require.NoError(t, err)
	transaction.ObjectVersion = 1
}

func getTransactionByID(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID, transactionID uuid.UUID) (*Transaction, error) {
	t.Helper()

	transaction := &Transaction{}
	err := db.QueryRow(ctx,
		`SELECT
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
		FROM transactions
		WHERE user_id = $1 AND id = $2;`,
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

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}

	return *id
}

func nullableUUIDValue(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}

	return id
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func ptr[T any](v T) *T {
	return &v
}
