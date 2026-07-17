package user

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("financial_tracker_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts("../../../database/schema.sql"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatal(err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err := testPool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	testPool.Close()
	postgresContainer.Terminate(ctx)
	os.Exit(code)
}

func TestPostgresUserRepository_Create(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresUserRepository(testPool)
	testUser := &User{Email: "test@gmail.com"}

	err := repo.Create(ctx, testUser)
	r.NoError(err)
	r.NotEqual(uuid.Nil, testUser.ID)
	r.Equal("test@gmail.com", testUser.Email)
	r.False(testUser.CreatedAt.IsZero())
	r.False(testUser.UpdatedAt.IsZero())
	r.Equal(int64(1), testUser.ObjectVersion)

	storedUser, err := getUserByID(t, ctx, testPool, testUser.ID)
	r.NoError(err)
	r.Equal(testUser.ID, storedUser.ID)
	r.Equal(testUser.Email, storedUser.Email)
	r.True(testUser.CreatedAt.Equal(storedUser.CreatedAt))
	r.True(testUser.UpdatedAt.Equal(storedUser.UpdatedAt))
	r.Equal(testUser.ObjectVersion, storedUser.ObjectVersion)
}

func TestPostgresUserRepository_Update(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresUserRepository(testPool)
	testUser := &User{Email: "test@gmail.com"}
	insertUser(t, ctx, testPool, testUser)
	prevVersion := testUser.ObjectVersion
	prevUpdatedAt := testUser.UpdatedAt

	testUser.Email = "updated@gmail.com"

	err := repo.Update(ctx, testUser)
	r.NoError(err)
	r.Equal("updated@gmail.com", testUser.Email)
	r.Equal(prevVersion+1, testUser.ObjectVersion)
	r.False(testUser.UpdatedAt.IsZero())
	r.True(testUser.UpdatedAt.After(prevUpdatedAt) || testUser.UpdatedAt.Equal(prevUpdatedAt))

	storedUser, err := getUserByID(t, ctx, testPool, testUser.ID)
	r.NoError(err)
	r.Equal(testUser.ID, storedUser.ID)
	r.Equal("updated@gmail.com", storedUser.Email)
	r.True(testUser.CreatedAt.Equal(storedUser.CreatedAt))
	r.True(testUser.UpdatedAt.Equal(storedUser.UpdatedAt))
	r.Equal(testUser.ObjectVersion, storedUser.ObjectVersion)
}

func TestPostgresUserRepository_Update_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresUserRepository(testPool)
	testUser := &User{
		ID:            uuid.New(),
		Email:         "missing@gmail.com",
		ObjectVersion: 1,
	}

	err := repo.Update(ctx, testUser)
	r.ErrorIs(err, ErrNotFound)
}

func TestPostgresUserRepository_GetByID(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &User{Email: "test@gmail.com"}
	insertUser(t, ctx, testPool, testUser)

	repo := NewPostgresUserRepository(testPool)
	storedUser, err := repo.GetByID(ctx, testUser.ID)
	r.NoError(err)
	r.Equal(testUser.ID, storedUser.ID)
	r.Equal(testUser.Email, storedUser.Email)
	r.True(testUser.CreatedAt.Equal(storedUser.CreatedAt))
	r.True(testUser.UpdatedAt.Equal(storedUser.UpdatedAt))
	r.Equal(testUser.ObjectVersion, storedUser.ObjectVersion)
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresUserRepository(testPool)
	user, err := repo.GetByID(ctx, uuid.New())
	r.ErrorIs(err, ErrNotFound)
	r.Nil(user)
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &User{Email: "test@gmail.com"}
	insertUser(t, ctx, testPool, testUser)

	repo := NewPostgresUserRepository(testPool)
	err := repo.Delete(ctx, testUser.ID)
	r.NoError(err)

	_, err = getUserByID(t, ctx, testPool, testUser.ID)
	r.Error(err)
	r.ErrorIs(err, pgx.ErrNoRows)
}

func TestPostgresUserRepository_Delete_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresUserRepository(testPool)
	err := repo.Delete(ctx, uuid.New())
	r.NoError(err)
}
