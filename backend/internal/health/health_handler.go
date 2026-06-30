package health

import (
	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	service *HealthService
}

func NewHealthHandler(service *HealthService) *HealthHandler {
	return &HealthHandler{
		service: service,
	}
}

func (hh *HealthHandler) Alive(c *gin.Context) {
	common.Ok(c, AliveResponse{Status: "ok"})
}

func (hh *HealthHandler) Ready(c *gin.Context) {
	readinesResponse := hh.service.Ready(c.Request.Context())
	common.Ok(c, readinesResponse)
}
