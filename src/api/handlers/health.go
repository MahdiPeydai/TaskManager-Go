package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/helpers"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health godoc
// Summary Health Check
// Description Health Check
// Tags health
// Produce Json
//
//	@Security	AuthBearer
//	@Success	200	{object}	helpers.BaseHTTPResponse
//	@Router		/v1/health [GET]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(200, helpers.GenerateBaseResponse("Working!", true, helpers.Success))
	return
}
