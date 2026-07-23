package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/cadezd/expense-tracker/internal/config"
	"github.com/cadezd/expense-tracker/internal/database"
	"github.com/cadezd/expense-tracker/internal/health"
	"github.com/cadezd/expense-tracker/internal/receipt"
	"github.com/cadezd/expense-tracker/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("uneptected error occoured: %v", err)
	}
}

func run() error {
	godotenv.Load()

	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, conf.DatabseURL)
	if err != nil {
		return fmt.Errorf("databse connection error: %v", err)
	}

	router := gin.Default()
	router.Use(common.ErrorHandler())

	v1 := router.Group("/api/v1")
	v1.Use(DevUserMiddleware(conf.TestUserID))
	{
		healthService := health.NewHealthService(pool)
		healthHandler := health.NewHealthHandler(healthService)
		v1.GET("/health/alive", healthHandler.Alive)
		v1.GET("/health/ready", healthHandler.Ready)

		receiptRepository := receipt.NewPostgresReceiptRepository(pool)
		receiptStorage := storage.NewLocalStorage(conf.UploadDirectory, conf.MaxFileUploadSizeInBytes)
		receiptService := receipt.NewReceiptService(receiptRepository, receiptStorage)
		receiptHandler := receipt.NewReceiptHandler(receiptService)
		v1.POST("/receipts", receiptHandler.Upload)
		v1.GET("/receipts", receiptHandler.List)
		v1.GET("/receipts/:id", receiptHandler.GetByID)
		v1.GET("/receipts/:id/file", receiptHandler.GetFile)
		v1.DELETE("/receipts/:id", receiptHandler.Delete)
	}

	router.Run()
	return nil
}

// TODO: remove when finished testing
func DevUserMiddleware(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}
