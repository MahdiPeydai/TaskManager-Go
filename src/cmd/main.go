package main

import (
	"github.com/mahdipeydai/taskmanager-go/api"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
)

var logger = logging.GetLogger(config.GetConfig())

// @contact.name				Mahdi Peydai
// @contact.email				mahdipeydai@gmail.com
//
// @title						TaskManager - Go
// @version					0.1
func main() {
	cfg := config.GetConfig()

	logger.Info(logging.Internal, logging.StartUp, "Setting up GIN server", nil)
	api.InitServer(cfg)
}
