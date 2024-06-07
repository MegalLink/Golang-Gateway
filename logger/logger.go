package logger

import (
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type IFastLogger interface {
	Debug(tag string, v interface{})
	Info(tag string, v interface{})
	Warning(tag string, v interface{})
	Error(tag string, v interface{})
	WithPrefix(prefix string)
}

type loggerOptions struct {
	tagDivider     string
	prefix         string
	dateTimeFormat string
}

type fastLogger struct {
	zLogger *zap.Logger
	loggerOptions
}

func NewFastLogger() (IFastLogger, error) {
	logger := &fastLogger{}
	logger.loggerOptions = loggerOptions{
		tagDivider:     " |",
		prefix:         "=>",
		dateTimeFormat: "02/01/2006 15:04:05",
	}

	zapConfig := logger.getZapConfig()
	zLogger, err := logger.getZapLogger(zapConfig)
	logger.zLogger = zLogger
	if err != nil {
		return logger, err
	}

	return logger, err
}

func (logger *fastLogger) getZapConfig() *zap.Config {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	config.Development = true
	config.EncoderConfig = zap.NewProductionEncoderConfig()
	config.EncoderConfig.ConsoleSeparator = " "
	config.EncoderConfig.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("%s%s%s", "(", level.CapitalString(), ")"))
	}
	config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(logger.dateTimeFormat))
		enc.AppendString(logger.prefix)
	}

	return &config
}

func (logger *fastLogger) getZapLogger(zapConfig *zap.Config) (*zap.Logger, error) {
	return zapConfig.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.FatalLevel+1),
	)
}

func (logger *fastLogger) Debug(tag string, v interface{}) {
	logger.withAnyData(v).Debug(tag + logger.tagDivider)
}

func (logger *fastLogger) Info(tag string, v interface{}) {
	logger.withAnyData(v).Info(tag + logger.tagDivider)
}
func (logger *fastLogger) Warning(tag string, v interface{}) {
	logger.withAnyData(v).Warn(tag + logger.tagDivider)
}
func (logger *fastLogger) Error(tag string, v interface{}) {
	logger.withAnyData(v).Error(tag + logger.tagDivider)
}

func (logger *fastLogger) withAnyData(i interface{}) *zap.Logger {
	//instead of just send i , we can implement a filter of i to mask fields
	return logger.zLogger.With(zap.Any("Data", i))
}

// New public function to set prefix dynamically
func (logger *fastLogger) WithPrefix(prefix string) {
	// Update the internal prefix value
	logger.prefix = prefix
	// Recreate the zap logger with the new prefix
	newZapConfig := logger.getZapConfig()
	newZLogger, err := logger.getZapLogger(newZapConfig)
	if err != nil {
		log.Panic(err)
	}
	logger.zLogger = newZLogger
}
