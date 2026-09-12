package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/handlers"
	"github.com/mahdipeydai/taskmanager-go/config"
)

func UsersRouter(routerGroup *gin.RouterGroup, cfg *config.Config) {
	h := handlers.GetUsersHandler(cfg)

	routerGroup.POST("/register", h.RegisterByUsername)
	routerGroup.POST("/login", h.LoginByUsername)
	routerGroup.POST("/refresh-token", h.RefreshToken)
}
