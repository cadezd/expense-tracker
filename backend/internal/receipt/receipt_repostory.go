package receipt

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("receipt not found")
)

type ReceiptRepository interface {
	Create(ctx context.Context, receipt *Receipt) error
	Update(ctx context.Context, receipt *Receipt) error
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
) error {
	sqlQuery := `
		INSERT INTO receipts (user_id, original_filename, stored_filename, storage_path, mime_type, file_size)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status, created_at, updated_at, object_version;
	`

	err := rr.db.QueryRow(ctx, sqlQuery,
		receipt.UserID,
		receipt.OriginalFilename,
		receipt.StoredFilename,
		receipt.StoragePath,
		receipt.MimeType,
		receipt.FileSize,
	).Scan(
		&receipt.ID,
		&receipt.Status,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
		&receipt.ObjectVersion,
	)
	if err != nil {
		return fmt.Errorf("create receipt error: %w", err)
	}

	return nil
}

func (rr *PostgresReceiptRepository) Update(
	ctx context.Context,
	receipt *Receipt,
) error {
	sqlQuery := `
		UPDATE receipts
		SET 
			original_filename = $1,
			stored_filename = $2,
			storage_path = $3,
			mime_type = $4,
			file_size = $5,
			status = $6,
			updated_at = NOW(),
			object_version = object_version + 1
		WHERE
			user_id = $7 AND
			id = $8 AND
			object_version = $9
		RETURNING
			updated_at,
			object_version;
	`

	err := rr.db.QueryRow(ctx, sqlQuery,
		receipt.OriginalFilename,
		receipt.StoredFilename,
		receipt.StoragePath,
		receipt.MimeType,
		receipt.FileSize,
		receipt.Status,
		receipt.UserID,
		receipt.ID,
		receipt.ObjectVersion,
	).Scan(
		&receipt.UpdatedAt,
		&receipt.ObjectVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update receipt error: %w", err)
	}

	return nil
}

func (rr *PostgresReceiptRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	receiptID uuid.UUID,
) (*Receipt, error) {
	receipt := &Receipt{}

	sqlQuery := `
		SELECT 
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
		WHERE 
			user_id = $1 AND 
			id = $2;
	`

	err := rr.db.QueryRow(ctx, sqlQuery,
		userID,
		receiptID,
	).Scan(
		&receipt.ID,
		&receipt.UserID,
		&receipt.OriginalFilename,
		&receipt.StoredFilename,
		&receipt.StoragePath,
		&receipt.MimeType,
		&receipt.FileSize,
		&receipt.Status,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
		&receipt.ObjectVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get receipt by id error: %w", err)
	}

	return receipt, nil
}

func (rr *PostgresReceiptRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	offset int,
	limit int,
) ([]*Receipt, error) {
	sqlQuery := `
		SELECT 
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
		WHERE 
			user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		OFFSET $3;
	`

	rows, err := rr.db.Query(ctx, sqlQuery,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list receipts error: %w", err)
	}
	defer rows.Close()

	receipts := make([]*Receipt, 0)
	for rows.Next() {
		receipt := &Receipt{}

		err := rows.Scan(
			&receipt.ID,
			&receipt.UserID,
			&receipt.OriginalFilename,
			&receipt.StoredFilename,
			&receipt.StoragePath,
			&receipt.MimeType,
			&receipt.FileSize,
			&receipt.Status,
			&receipt.CreatedAt,
			&receipt.UpdatedAt,
			&receipt.ObjectVersion,
		)
		if err != nil {
			return nil, fmt.Errorf("scan receipt error: %w", err)
		}

		receipts = append(receipts, receipt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate receipts error: %w", err)
	}

	return receipts, nil
}

func (rr *PostgresReceiptRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
	receiptID uuid.UUID,
) error {
	sqlQuery := `
		DELETE 
		FROM receipts 
		WHERE 
			user_id = $1 AND 
			id = $2;
	`

	_, err := rr.db.Exec(ctx, sqlQuery,
		userID,
		receiptID,
	)
	if err != nil {
		return fmt.Errorf("delete receipt error: %w", err)
	}

	return nil
}
