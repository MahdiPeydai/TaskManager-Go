package logging

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahdipeydai/golang-clean-web-api/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var zapLoggerLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

type zapLogger struct {
	config *config.Config
	logger *zap.SugaredLogger
}

func newZapLogger(cfg *config.Config) LoggerInterface {
	logger := &zapLogger{config: cfg}
	logger.Init()
	return logger
}

func (zl *zapLogger) getLoggerLevel() zapcore.Level {
	level, exist := zapLoggerLevelMap[zl.config.Logger.Level]
	if exist != true {
		level = zapcore.DebugLevel
	}
	return level
}

func (zl *zapLogger) Init() {
	fileName := fmt.Sprintf("%s%s-%s.%s", zl.config.Logger.FilePath, time.Now().Format("2006-01-02"), uuid.New(), "log")
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    1,
		MaxBackups: 10,
		MaxAge:     5,
		Compress:   true,
	})

	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(zapcore.NewJSONEncoder(zapConfig), writer, zl.getLoggerLevel())

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
	logger = logger.With("AppName", zl.config.Logger.AppName, "Logger", zl.config.Logger.LoggerName)

	zl.logger = logger
}

func (zl *zapLogger) Debug(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	params := prepareLogKeys(extras, cat, sub)
	zl.logger.Debugw(msg, params...)
}
func (zl *zapLogger) Debugf(template string, args ...interface{}) {
	zl.logger.Debugf(template, args...)
}

func (zl *zapLogger) Info(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	params := prepareLogKeys(extras, cat, sub)
	zl.logger.Infow(msg, params...)
}
func (zl *zapLogger) Infof(template string, args ...interface{}) {
	zl.logger.Infof(template, args...)
}

func (zl *zapLogger) Warn(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	params := prepareLogKeys(extras, cat, sub)
	zl.logger.Warnw(msg, params...)
}
func (zl *zapLogger) Warnf(template string, args ...interface{}) {
	zl.logger.Warnf(template, args...)
}

func (zl *zapLogger) Error(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	params := prepareLogKeys(extras, cat, sub)
	zl.logger.Errorw(msg, params...)
}
func (zl *zapLogger) Errorf(template string, args ...interface{}) {
	zl.logger.Errorf(template, args...)
}

func (zl *zapLogger) Fatal(cat LogCategory, sub LogSubCategory, msg string, extras map[ExtraKey]interface{}) {
	params := prepareLogKeys(extras, cat, sub)
	zl.logger.Fatalw(msg, params...)
}

func (zl *zapLogger) Fatalf(template string, args ...interface{}) {
	zl.logger.Fatalf(template, args...)
}

func prepareLogKeys(extras map[ExtraKey]interface{}, cat LogCategory, sub LogSubCategory) []interface{} {
	if extras == nil {
		extras = make(map[ExtraKey]interface{})
	}
	extras["category"] = string(cat)
	extras["subcategory"] = string(sub)

	return MapToZapParams(extras)
}
