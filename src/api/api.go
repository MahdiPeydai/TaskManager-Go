package api

import (
	"fmt"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/mahdipeydai/taskmanager-go/api/middlewares"
	"github.com/mahdipeydai/taskmanager-go/api/routers"
	"github.com/mahdipeydai/taskmanager-go/api/validators"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/docs"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitServer(cfg *config.Config) {
	engin := gin.New()
	registerCustomValidators(cfg)
	engin.Use(gin.Logger(), gin.CustomRecovery(middlewares.ErrorHandler))

	registerMiddlewares(engin, cfg)
	registerRoutes(engin, cfg)
	registerSwagger(engin, cfg)

	err := engin.Run(fmt.Sprintf(":%d", cfg.Server.InternalPort))
	if err != nil {
		return
	}
}

func registerMiddlewares(g *gin.Engine, cfg *config.Config) {
	g.Use(
		middlewares.LogRequestResponse(logging.GetLogger(cfg)),
		middlewares.Cors(cfg.Server.AllowOrigins),
		middlewares.LimitByRequest(float64(cfg.Server.RateLimit)),
	)
}

func registerRoutes(g *gin.Engine, cfg *config.Config) {
	apiGroup := g.Group("/api")
	v1Group := apiGroup.Group("/v1")

	{
		healthRouterGroup := v1Group.Group("/health")
		routers.HealthRouter(healthRouterGroup)

		usersRouterGroup := v1Group.Group("/users")
		routers.UsersRouter(usersRouterGroup, cfg)

		tasksRouterGroup := v1Group.Group("/tasks")
		routers.TasksRouter(tasksRouterGroup, cfg)
	}

}

func registerSwagger(g *gin.Engine, cfg *config.Config) {
	docs.SwaggerInfo.Description = "Swagger golang webservice API"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.Server.ExternalPort)
	docs.SwaggerInfo.Schemes = []string{"http"}

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func registerCustomValidators(cfg *config.Config) {
	val, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		err := val.RegisterValidation("userPassword", validators.UserPasswordValidator(cfg), true)
		if err != nil {
			extras := map[logging.ExtraKey]interface{}{
				logging.ErrorMessage: err.Error(),
			}
			logger.Fatal(logging.Internal, logging.StartUp, "Registering user password validator failed", extras)
		}
	}
}
