package user

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func resetDB(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(ctx, "TRUNCATE TABLE transactions, receipts, users RESTART IDENTITY CASCADE;")
	require.NoError(t, err)
}

func insertUser(t *testing.T, ctx context.Context, db *pgxpool.Pool, user *User) {
	t.Helper()

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	}
	if user.ObjectVersion == 0 {
		user.ObjectVersion = 1
	}

	err := db.QueryRow(ctx,
		`INSERT INTO users (
			id,
			email,
			created_at,
			updated_at,
			object_version
		)
		VALUES (
			COALESCE($1, gen_random_uuid()),
			$2,
			$3,
			$4,
			$5
		)
		RETURNING id, created_at, updated_at, object_version;`,
		nullableUUIDValue(user.ID),
		user.Email,
		user.CreatedAt,
		user.UpdatedAt,
		user.ObjectVersion,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.ObjectVersion,
	)
	require.NoError(t, err)
}

func getUserByID(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) (*User, error) {
	t.Helper()

	user := &User{}
	err := db.QueryRow(ctx,
		`SELECT id, email, created_at, updated_at, object_version
		 FROM users
		 WHERE id = $1;`,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.ObjectVersion,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return user, nil
}

func nullableUUIDValue(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}
