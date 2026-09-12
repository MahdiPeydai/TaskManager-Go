package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/handlers"
	"github.com/mahdipeydai/taskmanager-go/api/middlewares"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
)

func TasksRouter(routerGroup *gin.RouterGroup, cfg *config.Config) {
	h := handlers.GetTasksHandler(cfg)

	routerGroup.POST("/", middlewares.Authentication(cfg), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.Create)
	routerGroup.GET("/", middlewares.Authentication(cfg), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.GetByFilter)
	routerGroup.GET("/:id", middlewares.Authentication(cfg), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.GetByID)
	routerGroup.PUT("/:id", middlewares.Authentication(cfg), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.Update)
	routerGroup.DELETE("/:id", middlewares.Authentication(cfg), middlewares.Authorization([]string{constants.AdminRoleName, constants.DefaultRoleName}), h.Delete)
}
