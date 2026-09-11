package main

import (
	"github.com/mahdipeydai/golang-clean-web-api/api"
	"github.com/mahdipeydai/golang-clean-web-api/config"
	"github.com/mahdipeydai/golang-clean-web-api/pkg/logging"
)

var logger = logging.GetLogger(config.GetConfig())

func main() {
	cfg := config.GetConfig()

	logger.Info(logging.Internal, logging.StartUp, "Setting up GIN server", nil)
	api.InitServer(cfg)
}
