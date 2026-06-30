package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/cadezd/expense-tracker/internal/config"
	"github.com/cadezd/expense-tracker/internal/database"
	"github.com/cadezd/expense-tracker/internal/health"
	"github.com/gin-gonic/gin"
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
	{
		healthService := health.NewHealthService(pool)
		healthHandler := health.NewHealthHandler(healthService)
		v1.GET("/health/alive", healthHandler.Alive)
		v1.GET("/health/ready", healthHandler.Ready)
	}

	router.Run()
	return nil
}
