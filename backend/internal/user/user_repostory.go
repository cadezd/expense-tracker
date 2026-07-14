package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	Update(ctx context.Context, user User) (User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (User, error)
	List(ctx context.Context, offset, limit int) ([]User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}