package receipt

import (
	"context"
	"database/sql"
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

func insertReceipt(t *testing.T, ctx context.Context, db *pgxpool.Pool, receipt *Receipt) {
	t.Helper()

	if receipt.CreatedAt.IsZero() {
		receipt.CreatedAt = time.Now().UTC()
	}

	if receipt.Status == "" {
		receipt.Status = StatusUploaded
	}
	if receipt.UpdatedAt.IsZero() {
		receipt.UpdatedAt = receipt.CreatedAt
	}

	err := db.QueryRow(ctx,
		`INSERT INTO receipts (
			user_id,
			original_filename,
			stored_filename,
			storage_path,
			mime_type,
			file_size,
			status,
			created_at,
			updated_at,
			object_version
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id;`,
		receipt.UserID,
		receipt.OriginalFilename,
		receipt.StoredFilename,
		receipt.StoragePath,
		receipt.MimeType,
		receipt.FileSize,
		receipt.Status,
		receipt.CreatedAt,
		receipt.UpdatedAt,
		int64(1),
	).Scan(&receipt.ID)
	require.NoError(t, err)
	receipt.ObjectVersion = 1
}

func getReceiptByID(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID, receiptID uuid.UUID) (*Receipt, error) {
	t.Helper()

	receipt := &Receipt{}
	var fileSize sql.NullInt64

	err := db.QueryRow(ctx,
		`SELECT
			id,
			user_id,
			original_filename,
			stored_filename,
			storage_path,
			mime_type,
			file_size,
			status,
			created_at,
			updated_at,
			object_version
		FROM receipts
		WHERE user_id = $1 AND id = $2;`,
		userID,
		receiptID,
	).Scan(
		&receipt.ID,
		&receipt.UserID,
		&receipt.OriginalFilename,
		&receipt.StoredFilename,
		&receipt.StoragePath,
		&receipt.MimeType,
		&fileSize,
		&receipt.Status,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
		&receipt.ObjectVersion,
	)

	if err != nil {
		return nil, err
	}
	if fileSize.Valid {
		receipt.FileSize = &fileSize.Int64
	}

	return receipt, nil
}

func ptr[T any](v T) *T {
	return &v
}
