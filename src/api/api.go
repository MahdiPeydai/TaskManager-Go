package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/middlewares"
	"github.com/mahdipeydai/taskmanager-go/api/routers"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitServer(cfg *config.Config) {
	engin := gin.New()
	engin.Use(gin.Logger(), gin.CustomRecovery(middlewares.ErrorHandler))

	registerMiddlewares(engin, cfg)
	registerRoutes(engin)
	registerSwagger(engin, cfg)

	err := engin.Run(fmt.Sprintf(":%d", cfg.Server.InternalPort))
	if err != nil {
		return
	}
}

func registerMiddlewares(g *gin.Engine, cfg *config.Config) {
	g.Use(
		middlewares.Cors(cfg.Server.AllowOrigins),
		middlewares.LimitByRequest(float64(cfg.Server.RateLimit)),
	)
}

func registerRoutes(g *gin.Engine) {
	apiGroup := g.Group("/api")
	v1Group := apiGroup.Group("/v1")

	{
		healthRouterGroup := v1Group.Group("/health")
		routers.HealthRouter(healthRouterGroup)
	}

}

func registerSwagger(g *gin.Engine, cfg *config.Config) {
	docs.SwaggerInfo.Description = "Swagger golang webservice API"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.Server.ExternalPort)
	docs.SwaggerInfo.Schemes = []string{"http"}

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
