package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/handlers"
	"github.com/mahdipeydai/taskmanager-go/api/middlewares"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
)

func TasksRouter(routerGroup *gin.RouterGroup, cfg *config.Config, logger logging.LoggerInterface) {
	h := handlers.GetTasksHandler(cfg, logger)

	routerGroup.POST("", middlewares.Authentication(cfg, logger), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.Create)
	routerGroup.GET("", middlewares.Authentication(cfg, logger), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.GetByFilter)
	routerGroup.GET("/:id", middlewares.Authentication(cfg, logger), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.GetByID)
	routerGroup.PUT("/:id", middlewares.Authentication(cfg, logger), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.Update)
	routerGroup.DELETE("/:id", middlewares.Authentication(cfg, logger), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.Delete)
}
