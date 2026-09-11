package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/golang-clean-web-api/api/helpers"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(200, helpers.GenerateBaseResponse("Working!", true, helpers.Success))
	return
}
