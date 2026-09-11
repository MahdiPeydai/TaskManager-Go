package logging

import (
	"sync"

	"github.com/mahdipeydai/golang-clean-web-api/config"
)

type LoggerInterface interface {
	Init()
	Debug(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{})
	Debugf(template string, args ...interface{})
	Info(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{})
	Infof(template string, args ...interface{})
	Warn(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{})
	Warnf(template string, args ...interface{})
	Error(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{})
	Errorf(template string, args ...interface{})
	Fatal(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{})
	Fatalf(template string, args ...interface{})
}

var once sync.Once
var logger LoggerInterface

// GetLogger implementing singleton
func GetLogger(cfg *config.Config) LoggerInterface {
	once.Do(func() {
		if cfg.Logger.LoggerName == "zap" {
			logger = newZapLogger(cfg)
		} else if cfg.Logger.LoggerName == "zero" {
			logger = newZeroLogger(cfg)
		} else {
			panic("Logger not supported")
		}
	})
	return logger
}
