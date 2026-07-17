package transaction

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cadezd/expense-tracker/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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

func TestTransactionRepository_Create(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{Email: "test@gmail.com"}
	seedUser(t, ctx, testPool, testUser)

	repo := NewPostgresTransactionRepository(testPool)
	counterparty := "Coffee Shop"
	description := "Morning coffee"
	txn := &Transaction{
		UserID:          testUser.ID,
		Type:            TypeExpense,
		Counterparty:    &counterparty,
		Amount:          decimal.RequireFromString("12.34"),
		Currency:        "EUR",
		TransactionDate: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
		Category:        "food",
		Description:     &description,
	}

	err := repo.Create(ctx, txn)
	r.NoError(err)
	r.NotEqual(uuid.Nil, txn.ID)
	r.False(txn.CreatedAt.IsZero())
	r.False(txn.UpdatedAt.IsZero())
	r.Equal(int64(1), txn.ObjectVersion)
	r.Nil(txn.ReceiptID)
	r.NotNil(txn.Counterparty)
	r.Equal(counterparty, *txn.Counterparty)
	r.NotNil(txn.Description)
	r.Equal(description, *txn.Description)

	storedTxn, err := getTransactionByID(t, ctx, testPool, testUser.ID, txn.ID)
	r.NoError(err)
	r.Equal(txn.ID, storedTxn.ID)
	r.Equal(txn.UserID, storedTxn.UserID)
	r.Nil(storedTxn.ReceiptID)
	r.Equal(txn.Type, storedTxn.Type)
	r.NotNil(storedTxn.Counterparty)
	r.Equal(counterparty, *storedTxn.Counterparty)
	r.True(storedTxn.Amount.Equal(txn.Amount))
	r.Equal(txn.Currency, storedTxn.Currency)
	r.Equal(txn.TransactionDate.Format("2006-01-02"), storedTxn.TransactionDate.Format("2006-01-02"))
	r.Equal(txn.Category, storedTxn.Category)
	r.NotNil(storedTxn.Description)
	r.Equal(description, *storedTxn.Description)
	r.Equal(int64(1), storedTxn.ObjectVersion)
}

func TestTransactionRepository_Update(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{Email: "test@gmail.com"}
	seedUser(t, ctx, testPool, testUser)

	repo := NewPostgresTransactionRepository(testPool)
	initialCounterparty := "Coffee Shop"
	initialDescription := "Morning coffee"
	txn := &Transaction{
		UserID:          testUser.ID,
		Type:            TypeExpense,
		Counterparty:    &initialCounterparty,
		Amount:          decimal.RequireFromString("12.34"),
		Currency:        "EUR",
		TransactionDate: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
		Category:        "food",
		Description:     &initialDescription,
		CreatedAt:       time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
	}
	insertTransaction(t, ctx, testPool, txn)
	prevVersion := txn.ObjectVersion
	prevUpdatedAt := txn.UpdatedAt

	updatedCounterparty := "Lunch Spot"
	updatedDescription := "Lunch with team"
	txn.Counterparty = &updatedCounterparty
	txn.Amount = decimal.RequireFromString("18.90")
	txn.Currency = "USD"
	txn.TransactionDate = time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)
	txn.Category = "meals"
	txn.Description = &updatedDescription

	err := repo.Update(ctx, txn)
	r.NoError(err)
	r.Equal(prevVersion+1, txn.ObjectVersion)
	r.False(txn.UpdatedAt.IsZero())
	r.True(txn.UpdatedAt.After(prevUpdatedAt) || txn.UpdatedAt.Equal(prevUpdatedAt))
	r.NotNil(txn.Counterparty)
	r.Equal(updatedCounterparty, *txn.Counterparty)
	r.True(txn.Amount.Equal(decimal.RequireFromString("18.90")))
	r.Equal("USD", txn.Currency)
	r.Equal("2024-01-03", txn.TransactionDate.Format("2006-01-02"))
	r.Equal("meals", txn.Category)
	r.NotNil(txn.Description)
	r.Equal(updatedDescription, *txn.Description)

	storedTxn, err := getTransactionByID(t, ctx, testPool, testUser.ID, txn.ID)
	r.NoError(err)
	r.Equal(txn.ID, storedTxn.ID)
	r.Equal(txn.UserID, storedTxn.UserID)
	r.Nil(storedTxn.ReceiptID)
	r.Equal(txn.Type, storedTxn.Type)
	r.NotNil(storedTxn.Counterparty)
	r.Equal(updatedCounterparty, *storedTxn.Counterparty)
	r.True(storedTxn.Amount.Equal(txn.Amount))
	r.Equal(txn.Currency, storedTxn.Currency)
	r.Equal(txn.TransactionDate.Format("2006-01-02"), storedTxn.TransactionDate.Format("2006-01-02"))
	r.Equal(txn.Category, storedTxn.Category)
	r.NotNil(storedTxn.Description)
	r.Equal(updatedDescription, *storedTxn.Description)
	r.Equal(txn.ObjectVersion, storedTxn.ObjectVersion)
}

func TestTransactionRepository_Update_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresTransactionRepository(testPool)
	txn := &Transaction{
		UserID:          uuid.New(),
		Type:            TypeExpense,
		Amount:          decimal.RequireFromString("1.00"),
		Currency:        "EUR",
		TransactionDate: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
		Category:        "misc",
		ObjectVersion:   1,
	}

	err := repo.Update(ctx, txn)
	r.ErrorIs(err, ErrNotFound)
}

func TestTransactionRepository_GetByID(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{Email: "test@gmail.com"}
	seedUser(t, ctx, testPool, testUser)

	counterparty := "Coffee Shop"
	description := "Morning coffee"
	txn := &Transaction{
		UserID:          testUser.ID,
		Type:            TypeExpense,
		Counterparty:    &counterparty,
		Amount:          decimal.RequireFromString("12.34"),
		Currency:        "EUR",
		TransactionDate: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
		Category:        "food",
		Description:     &description,
		CreatedAt:       time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
	}
	insertTransaction(t, ctx, testPool, txn)

	repo := NewPostgresTransactionRepository(testPool)
	storedTxn, err := repo.GetByID(ctx, testUser.ID, txn.ID)
	r.NoError(err)
	r.Equal(txn.ID, storedTxn.ID)
	r.Equal(txn.UserID, storedTxn.UserID)
	r.Nil(storedTxn.ReceiptID)
	r.Equal(txn.Type, storedTxn.Type)
	r.NotNil(storedTxn.Counterparty)
	r.Equal(counterparty, *storedTxn.Counterparty)
	r.True(storedTxn.Amount.Equal(txn.Amount))
	r.Equal(txn.Currency, storedTxn.Currency)
	r.Equal(txn.TransactionDate.Format("2006-01-02"), storedTxn.TransactionDate.Format("2006-01-02"))
	r.Equal(txn.Category, storedTxn.Category)
	r.NotNil(storedTxn.Description)
	r.Equal(description, *storedTxn.Description)
	r.Equal(int64(1), storedTxn.ObjectVersion)
}

func TestTransactionRepository_GetByID_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresTransactionRepository(testPool)
	txn, err := repo.GetByID(ctx, uuid.New(), uuid.New())
	r.ErrorIs(err, ErrNotFound)
	r.Nil(txn)
}

func TestTransactionRepository_List(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	targetUser := &user.User{Email: "target@gmail.com"}
	seedUser(t, ctx, testPool, targetUser)

	otherUser := &user.User{Email: "other@gmail.com"}
	seedUser(t, ctx, testPool, otherUser)

	repo := NewPostgresTransactionRepository(testPool)

	baseDate := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	older := time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC)
	middle := time.Date(2024, time.January, 1, 11, 0, 0, 0, time.UTC)
	newer := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	otherTime := time.Date(2024, time.January, 1, 13, 0, 0, 0, time.UTC)

	txnA := &Transaction{
		UserID:          targetUser.ID,
		Type:            TypeExpense,
		Counterparty:    ptr("Bakery"),
		Amount:          decimal.RequireFromString("3.50"),
		Currency:        "EUR",
		TransactionDate: baseDate,
		Category:        "food",
		Description:     ptr("Bread"),
		CreatedAt:       older,
	}
	insertTransaction(t, ctx, testPool, txnA)

	txnB := &Transaction{
		UserID:          targetUser.ID,
		Type:            TypeExpense,
		Counterparty:    ptr("Cafe"),
		Amount:          decimal.RequireFromString("6.00"),
		Currency:        "EUR",
		TransactionDate: baseDate,
		Category:        "food",
		Description:     ptr("Coffee"),
		CreatedAt:       middle,
	}
	insertTransaction(t, ctx, testPool, txnB)

	txnC := &Transaction{
		UserID:          targetUser.ID,
		Type:            TypeExpense,
		Counterparty:    ptr("Market"),
		Amount:          decimal.RequireFromString("12.00"),
		Currency:        "EUR",
		TransactionDate: baseDate,
		Category:        "groceries",
		Description:     ptr("Snacks"),
		CreatedAt:       newer,
	}
	insertTransaction(t, ctx, testPool, txnC)

	otherTxn := &Transaction{
		UserID:          otherUser.ID,
		Type:            TypeIncome,
		Counterparty:    ptr("Employer"),
		Amount:          decimal.RequireFromString("1000.00"),
		Currency:        "EUR",
		TransactionDate: baseDate,
		Category:        "salary",
		Description:     ptr("Payday"),
		CreatedAt:       otherTime,
	}
	insertTransaction(t, ctx, testPool, otherTxn)

	transactions, err := repo.List(ctx, targetUser.ID, 1, 2)
	r.NoError(err)
	r.Len(transactions, 2)

	r.Equal(txnB.ID, transactions[0].ID)
	r.Equal(txnA.ID, transactions[1].ID)
	r.Equal(targetUser.ID, transactions[0].UserID)
	r.Equal(targetUser.ID, transactions[1].UserID)
	r.Equal("Cafe", *transactions[0].Counterparty)
	r.Equal("Bakery", *transactions[1].Counterparty)
	r.True(transactions[0].Amount.Equal(txnB.Amount))
	r.True(transactions[1].Amount.Equal(txnA.Amount))
	r.Equal("food", transactions[0].Category)
	r.Equal("food", transactions[1].Category)
}

func TestTransactionRepository_Delete(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	testUser := &user.User{Email: "test@gmail.com"}
	seedUser(t, ctx, testPool, testUser)

	repo := NewPostgresTransactionRepository(testPool)
	txn := &Transaction{
		UserID:          testUser.ID,
		Type:            TypeExpense,
		Counterparty:    ptr("Coffee Shop"),
		Amount:          decimal.RequireFromString("12.34"),
		Currency:        "EUR",
		TransactionDate: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
		Category:        "food",
		Description:     ptr("Morning coffee"),
		CreatedAt:       time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC),
	}
	insertTransaction(t, ctx, testPool, txn)

	err := repo.Delete(ctx, testUser.ID, txn.ID)
	r.NoError(err)

	_, err = getTransactionByID(t, ctx, testPool, testUser.ID, txn.ID)
	r.Error(err)
	r.ErrorIs(err, pgx.ErrNoRows)
}

func TestTransactionRepository_Delete_NotFound(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	resetDB(t, ctx, testPool)

	repo := NewPostgresTransactionRepository(testPool)
	err := repo.Delete(ctx, uuid.New(), uuid.New())
	r.NoError(err)
}
