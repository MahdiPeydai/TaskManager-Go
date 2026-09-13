package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/handlers"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
)

func UsersRouter(routerGroup *gin.RouterGroup, cfg *config.Config, logger logging.LoggerInterface) {
	h := handlers.GetUsersHandler(cfg, logger)

	routerGroup.POST("/register", h.RegisterByUsername)
	routerGroup.POST("/login", h.LoginByUsername)
	routerGroup.POST("/refresh-token", h.RefreshToken)
}
