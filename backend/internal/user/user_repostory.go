package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	List(ctx context.Context, offset, limit int) ([]*User, error)
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
) (*User, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (ur *PostgresUserRepository) Update(
	ctx context.Context,
	user *User,
) (*User, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (ur *PostgresUserRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
) (*User, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (ur *PostgresUserRepository) List(
	ctx context.Context,
	offset int,
	limit int,
) ([]*User, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (ur *PostgresUserRepository) Delete(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return fmt.Errorf("Not implemented")
}
