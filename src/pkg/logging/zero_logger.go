package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var zeroLoggerLevelMap = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
}

type zeroLogger struct {
	config *config.Config
	logger zerolog.Logger
}

func newZeroLogger(cfg *config.Config) *zeroLogger {
	logger := zeroLogger{config: cfg}
	logger.Init()
	return &logger
}

func (zl *zeroLogger) Init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimestampFieldName = "ts"
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000-0700"

	fileName := fmt.Sprintf("%s%s-%s.%s", zl.config.Logger.FilePath, time.Now().Format("2006-01-02"), uuid.New(), "log")

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(file).
		With().
		Timestamp().
		Str("AppName", zl.config.Logger.AppName).
		Str("Logger", zl.config.Logger.LoggerName).
		Logger()

	zerolog.SetGlobalLevel(zl.getLoggerLevel())

	zl.logger = logger
}

func (zl *zeroLogger) Debug(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	zl.logger.Debug().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(MapToZeroParams(extras)).
		Msg(msg)
}
func (zl *zeroLogger) Debugf(template string, args ...interface{}) {
	zl.logger.Debug().Msgf(template, args...)
}

func (zl *zeroLogger) Info(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	zl.logger.Info().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(MapToZeroParams(extras)).
		Msg(msg)
}
func (zl *zeroLogger) Infof(template string, args ...interface{}) {
	zl.logger.Info().Msgf(template, args...)
}

func (zl *zeroLogger) Warn(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	zl.logger.Warn().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(MapToZeroParams(extras)).
		Msg(msg)
}
func (zl *zeroLogger) Warnf(template string, args ...interface{}) {
	zl.logger.Warn().Msgf(template, args...)
}

func (zl *zeroLogger) Error(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	zl.logger.Error().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(MapToZeroParams(extras)).
		Msg(msg)
}
func (zl *zeroLogger) Errorf(template string, args ...interface{}) {
	zl.logger.Error().Msgf(template, args...)
}

func (zl *zeroLogger) Fatal(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	zl.logger.Fatal().
		Str("Category", string(cat)).
		Str("SubCategory", string(sub)).
		Fields(MapToZeroParams(extras)).
		Msg(msg)
}
func (zl *zeroLogger) Fatalf(template string, args ...interface{}) {
	zl.logger.Fatal().Msgf(template, args...)
}

func (zl *zeroLogger) getLoggerLevel() zerolog.Level {
	level, exist := zeroLoggerLevelMap[zl.config.Logger.Level]
	if !exist {
		level = zerolog.DebugLevel
	}
	return level
}
