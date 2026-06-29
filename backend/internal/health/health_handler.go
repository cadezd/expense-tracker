package health

import (
	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Alive(c *gin.Context) {
	common.Ok(c, AliveResponse{
		Status: "ok",
	})
}
