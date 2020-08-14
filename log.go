package gorm_admin

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger
var (
	Debug, Warn, Info, Error, DPanic, Panic, Fatal func(template string, args ...interface{})
)

func ConfigZapLog(logLevel string) {
	var allCore []zapcore.Core

	level := zap.DebugLevel
	if logLevel == "INFO" {
		level = zap.InfoLevel
	} else if logLevel == "WARN" {
		level = zap.WarnLevel
	} else if logLevel == "ERROR" {
		level = zap.ErrorLevel
	}
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05 "))
	}
	allCore = append(allCore, zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		level),
	)

	core := zapcore.NewTee(allCore...)
	logger = zap.New(core).WithOptions(zap.AddCaller()).Sugar()

	Debug = logger.Debugf
	Warn = logger.Warnf
	Info = logger.Infof
	Error = logger.Errorf
	DPanic = logger.DPanicf
	Panic = logger.Panicf
	Fatal = logger.Fatalf
}
