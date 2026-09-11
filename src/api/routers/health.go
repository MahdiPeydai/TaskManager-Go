package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/golang-clean-web-api/api/handlers"
)

func HealthRouter(gr *gin.RouterGroup) {
	handler := handlers.NewHealthHandler()
	gr.GET("/", handler.Health)
}
