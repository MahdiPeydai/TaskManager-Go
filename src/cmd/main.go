package main

import (
	"context"

	"github.com/mahdipeydai/taskmanager-go/api"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/data/cache"
	"github.com/mahdipeydai/taskmanager-go/data/db"
	migration "github.com/mahdipeydai/taskmanager-go/data/db/migration"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/pkg/tracing"
)

var logger = logging.GetLogger(config.GetConfig())

// @contact.name				Mahdi Peydai
// @contact.email				mahdipeydai@gmail.com
//
// @title						TaskManager - Go
// @version					0.1
//
// @securityDefinitions.apiKey	AuthBearer
// @in							header
// @name						Authorization
func main() {
	cfg := config.GetConfig()

	shutdownTracing, err := tracing.Init(context.Background(), cfg)
	if err != nil {
		logger.Fatal(logging.Opentelemetry, logging.StartUp, err.Error(), nil)
	}
	defer shutdownTracing(context.Background())

	err = db.InitDb(cfg)
	if err != nil {
		logger.Fatal(logging.Postgres, logging.StartUp, err.Error(), nil)
	}
	defer db.CloseDb()
	migration.UpInit()

	err = cache.InitRedis(cfg)
	if err != nil {
		logger.Fatal(logging.Redis, logging.StartUp, err.Error(), nil)
	}
	defer cache.CloseRedis(cfg)

	logger.Info(logging.Internal, logging.StartUp, "Setting up GIN server", nil)
	api.InitServer(cfg)
}
