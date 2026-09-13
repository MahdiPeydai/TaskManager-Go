package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/handlers"
)

func HealthRouter(gr *gin.RouterGroup) {
	handler := handlers.NewHealthHandler()
	gr.GET("", handler.Health)
}
