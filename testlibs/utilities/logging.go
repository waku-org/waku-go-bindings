package utilities

import (
	"go.uber.org/zap"
	
)

var devLogger *zap.Logger

func init() {
	var err error
	devLogger, err = zap.NewDevelopment()
	if err != nil {
		panic("Failed to create development logger: " + err.Error())
	}
}

func Debug(msg string, fields ...zap.Field) {
	devLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	devLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	devLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	devLogger.Error(msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	devLogger.DPanic(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	devLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	devLogger.Fatal(msg, fields...)
}

func Sync() {
	_ = devLogger.Sync()
}
