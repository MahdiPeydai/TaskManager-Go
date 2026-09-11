package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/golang-clean-web-api/api/middlewares"
	"github.com/mahdipeydai/golang-clean-web-api/api/routers"
	"github.com/mahdipeydai/golang-clean-web-api/config"
)

func InitServer(cfg *config.Config) {
	engin := gin.New()
	engin.Use(gin.Logger(), gin.CustomRecovery(middlewares.ErrorHandler))

	registerRoutes(engin)

	err := engin.Run(fmt.Sprintf(":%d", cfg.Server.InternalPort))
	if err != nil {
		return
	}
}

func registerRoutes(g *gin.Engine) {
	apiGroup := g.Group("/api")
	v1Group := apiGroup.Group("/v1")

	{
		healthRouterGroup := v1Group.Group("/health")
		routers.HealthRouter(healthRouterGroup)
	}

}
