package main

import (
	"github.com/mahdipeydai/taskmanager-go/api"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/data/db"
	migration "github.com/mahdipeydai/taskmanager-go/data/db/migration"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
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

	err := db.InitDb(cfg)
	if err != nil {
		logger.Fatal(logging.Postgres, logging.StartUp, err.Error(), nil)
	}
	defer db.CloseDb()
	migration.UpInit()

	logger.Info(logging.Internal, logging.StartUp, "Setting up GIN server", nil)
	api.InitServer(cfg)
}
