package main

import (
	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/cadezd/expense-tracker/internal/health"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(common.ErrorHandler())

	v1 := router.Group("/api/v1")
	{
		healthHandler := health.NewHealthHandler()
		v1.GET("/health/alive", healthHandler.Alive)
	}

	router.Run()
}
