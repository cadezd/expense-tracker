package receipt

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/cadezd/expense-tracker/internal/user"
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

func TestPostgresReceiptRepository_Create(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{
		Email: "test@gmail.com",
	}
	seedUser(t, ctx, testPool, testUser)

	rr := NewPostgresReceiptRepository(testPool)

	testReceipt := &Receipt{
		UserID:           testUser.ID,
		OriginalFilename: "img0001.jpeg",
		StoredFilename:   "img1234.jpeg",
		StoragePath:      "/uploads/img1234.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(123)),
	}
	err := rr.Create(ctx, testReceipt)
	r.NoError(err)
	r.NotEqual(uuid.Nil, testReceipt.ID)
	r.Equal(StatusUploaded, testReceipt.Status)
	r.Equal(int64(1), testReceipt.ObjectVersion)

	storedReceiptFromDB, err := getReceiptByID(t, ctx, testPool, testUser.ID, testReceipt.ID)
	r.NoError(err)
	r.Equal(testReceipt.ID, storedReceiptFromDB.ID)
	r.Equal(testReceipt.UserID, storedReceiptFromDB.UserID)
	r.Equal(testReceipt.OriginalFilename, storedReceiptFromDB.OriginalFilename)
	r.Equal(testReceipt.StoredFilename, storedReceiptFromDB.StoredFilename)
	r.Equal(testReceipt.StoragePath, storedReceiptFromDB.StoragePath)
	r.Equal(testReceipt.MimeType, storedReceiptFromDB.MimeType)
	r.Equal(StatusUploaded, storedReceiptFromDB.Status)
	r.NotNil(storedReceiptFromDB.FileSize)
	r.Equal(int64(123), *storedReceiptFromDB.FileSize)
}

func TestPostgresReceiptRepository_Update(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{Email: "test@gmail.com"}
	seedUser(t, ctx, testPool, testUser)

	repo := NewPostgresReceiptRepository(testPool)
	testReceipt := &Receipt{
		UserID:           testUser.ID,
		OriginalFilename: "img0001.jpeg",
		StoredFilename:   "img1234.jpeg",
		StoragePath:      "/uploads/img1234.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(123)),
	}
	insertReceipt(t, ctx, testPool, testReceipt)
	prevVersion := testReceipt.ObjectVersion
	prevUpdatedAt := testReceipt.UpdatedAt

	testReceipt.OriginalFilename = "img0002.jpeg"
	testReceipt.StoredFilename = "img4321.jpeg"
	testReceipt.StoragePath = "/uploads/img4321.jpeg"
	testReceipt.MimeType = "image/png"
	testReceipt.FileSize = ptr(int64(456))
	testReceipt.Status = StatusProcessed

	err := repo.Update(ctx, testReceipt)
	r.NoError(err)
	r.NotEqual(uuid.Nil, testReceipt.ID)
	r.Equal(testUser.ID, testReceipt.UserID)
	r.Equal("img0002.jpeg", testReceipt.OriginalFilename)
	r.Equal("img4321.jpeg", testReceipt.StoredFilename)
	r.Equal("/uploads/img4321.jpeg", testReceipt.StoragePath)
	r.Equal("image/png", testReceipt.MimeType)
	r.NotNil(testReceipt.FileSize)
	r.Equal(int64(456), *testReceipt.FileSize)
	r.Equal(StatusProcessed, testReceipt.Status)
	r.Equal(prevVersion+1, testReceipt.ObjectVersion)
	r.False(testReceipt.UpdatedAt.IsZero())
	r.True(testReceipt.UpdatedAt.After(prevUpdatedAt) || testReceipt.UpdatedAt.Equal(prevUpdatedAt))

	storedReceiptFromDB, err := getReceiptByID(t, ctx, testPool, testUser.ID, testReceipt.ID)
	r.NoError(err)
	r.Equal(testReceipt.OriginalFilename, storedReceiptFromDB.OriginalFilename)
	r.Equal(testReceipt.StoredFilename, storedReceiptFromDB.StoredFilename)
	r.Equal(testReceipt.StoragePath, storedReceiptFromDB.StoragePath)
	r.Equal(testReceipt.MimeType, storedReceiptFromDB.MimeType)
	r.NotNil(storedReceiptFromDB.FileSize)
	r.Equal(int64(456), *storedReceiptFromDB.FileSize)
	r.Equal(StatusProcessed, storedReceiptFromDB.Status)
	r.Equal(testReceipt.ObjectVersion, storedReceiptFromDB.ObjectVersion)
}

func TestPostgresReceiptRepository_GetByID(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{
		Email: "test@gmail.com",
	}
	seedUser(t, ctx, testPool, testUser)

	testReceipt := &Receipt{
		UserID:           testUser.ID,
		OriginalFilename: "img0001.jpeg",
		StoredFilename:   "img1234.jpeg",
		StoragePath:      "/uploads/img1234.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(123)),
	}
	insertReceipt(t, ctx, testPool, testReceipt)

	rr := NewPostgresReceiptRepository(testPool)

	storedReceiptFromDB, err := rr.GetByID(ctx, testUser.ID, testReceipt.ID)
	r.NoError(err)

	r.Equal(testReceipt.ID, storedReceiptFromDB.ID)
	r.Equal(testReceipt.UserID, storedReceiptFromDB.UserID)
	r.Equal(testReceipt.OriginalFilename, storedReceiptFromDB.OriginalFilename)
	r.Equal(testReceipt.StoredFilename, storedReceiptFromDB.StoredFilename)
	r.Equal(testReceipt.StoragePath, storedReceiptFromDB.StoragePath)
	r.Equal(testReceipt.MimeType, storedReceiptFromDB.MimeType)
	r.Equal(StatusUploaded, storedReceiptFromDB.Status)
	r.NotNil(storedReceiptFromDB.FileSize)
	r.Equal(int64(123), *storedReceiptFromDB.FileSize)
}

func TestPostgresReceiptRepository_GetByID_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresReceiptRepository(testPool)
	receipt, err := repo.GetByID(ctx, uuid.New(), uuid.New())
	r.ErrorIs(err, common.ErrNotFound)
	r.Nil(receipt)
}

func TestPostgresReceiptRepository_List(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	targetUser := &user.User{Email: "target@gmail.com"}
	seedUser(t, ctx, testPool, targetUser)

	otherUser := &user.User{Email: "other@gmail.com"}
	seedUser(t, ctx, testPool, otherUser)

	repo := NewPostgresReceiptRepository(testPool)

	older := time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC)
	middle := time.Date(2024, time.January, 1, 11, 0, 0, 0, time.UTC)
	newer := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	otherTime := time.Date(2024, time.January, 1, 13, 0, 0, 0, time.UTC)

	receiptA := &Receipt{
		UserID:           targetUser.ID,
		OriginalFilename: "a.jpeg",
		StoredFilename:   "a-stored.jpeg",
		StoragePath:      "/uploads/a-stored.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(100)),
		CreatedAt:        older,
	}
	insertReceipt(t, ctx, testPool, receiptA)

	receiptB := &Receipt{
		UserID:           targetUser.ID,
		OriginalFilename: "b.jpeg",
		StoredFilename:   "b-stored.jpeg",
		StoragePath:      "/uploads/b-stored.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(200)),
		CreatedAt:        middle,
	}
	insertReceipt(t, ctx, testPool, receiptB)

	receiptC := &Receipt{
		UserID:           targetUser.ID,
		OriginalFilename: "c.jpeg",
		StoredFilename:   "c-stored.jpeg",
		StoragePath:      "/uploads/c-stored.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(300)),
		CreatedAt:        newer,
	}
	insertReceipt(t, ctx, testPool, receiptC)

	otherReceipt := &Receipt{
		UserID:           otherUser.ID,
		OriginalFilename: "other.jpeg",
		StoredFilename:   "other-stored.jpeg",
		StoragePath:      "/uploads/other-stored.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(400)),
		CreatedAt:        otherTime,
	}
	insertReceipt(t, ctx, testPool, otherReceipt)

	receipts, err := repo.List(ctx, targetUser.ID, 1, 2)
	r.NoError(err)
	r.Len(receipts, 2)

	r.Equal(receiptB.ID, receipts[0].ID)
	r.Equal(receiptA.ID, receipts[1].ID)
	r.Equal(targetUser.ID, receipts[0].UserID)
	r.Equal(targetUser.ID, receipts[1].UserID)
	r.Equal("b.jpeg", receipts[0].OriginalFilename)
	r.Equal("a.jpeg", receipts[1].OriginalFilename)
	r.Equal(StatusUploaded, receipts[0].Status)
	r.Equal(StatusUploaded, receipts[1].Status)
}

func TestPostgresReceiptRepository_Delete(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{
		Email: "test@gmail.com",
	}
	seedUser(t, ctx, testPool, testUser)

	testReceipt := &Receipt{
		UserID:           testUser.ID,
		OriginalFilename: "img0001.jpeg",
		StoredFilename:   "img1234.jpeg",
		StoragePath:      "/uploads/img1234.jpeg",
		MimeType:         "image/jpeg",
		FileSize:         ptr(int64(123)),
	}
	insertReceipt(t, ctx, testPool, testReceipt)

	repo := NewPostgresReceiptRepository(testPool)

	err := repo.Delete(ctx, testUser.ID, testReceipt.ID)
	r.NoError(err)

	_, err = getReceiptByID(t, ctx, testPool, testUser.ID, testReceipt.ID)
	r.Error(err)
	r.ErrorIs(err, pgx.ErrNoRows)
}

func TestPostgresReceiptRepository_Delete_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)
	repo := NewPostgresReceiptRepository(testPool)
	err := repo.Delete(ctx, uuid.New(), uuid.New())
	r.NoError(err)
}
