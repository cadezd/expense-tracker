package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("user not found")
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

func (ur *PostgresUserRepository) Create(
	ctx context.Context,
	user *User,
) error {
	query := `
		INSERT INTO users (email)
		VALUES ($1)
		RETURNING id, created_at, updated_at, object_version;
	`

	err := ur.db.QueryRow(ctx, query, user.Email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.ObjectVersion,
	)
	if err != nil {
		return fmt.Errorf("create user error: %w", err)
	}

	return nil
}

func (ur *PostgresUserRepository) Update(
	ctx context.Context,
	user *User,
) error {
	query := `
		UPDATE users
		SET
			email = $1,
			updated_at = NOW(),
			object_version = object_version + 1
		WHERE
			id = $2 AND
			object_version = $3
		RETURNING updated_at, object_version;
	`

	err := ur.db.QueryRow(ctx, query,
		user.Email,
		user.ID,
		user.ObjectVersion,
	).Scan(
		&user.UpdatedAt,
		&user.ObjectVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("update user error: %w", err)
	}

	return nil
}

func (ur *PostgresUserRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
) (*User, error) {
	user := &User{}

	query := `
		SELECT id, email, created_at, updated_at, object_version
		FROM users
		WHERE id = $1;
	`

	err := ur.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.ObjectVersion,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id error: %w", err)
	}

	return user, nil
}

func (ur *PostgresUserRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
) error {
	_, err := ur.db.Exec(ctx, `DELETE FROM users WHERE id = $1;`, userID)
	if err != nil {
		return fmt.Errorf("delete user error: %w", err)
	}

	return nil
}
