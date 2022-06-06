package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog zap.Logger

func New() error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339) // Time readable to the user "2022-05-18T20:31:33+02:00"
	logger, err := config.Build()
	zapLog = *logger

	return err
}

func Get() zap.Logger {
	return zapLog
}

func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

func Error(message string, err error) {
	zapLog.Error(message, zap.Error(err))
}

func Fatal(message string, err error) {
	zapLog.Fatal(message, zap.Error(err))
}

func HttpWarn(message string, method zapcore.Field, uri zapcore.Field, status zapcore.Field) {
	zapLog.Info(message, method, uri, status)
}

func HttpError(message string, method zapcore.Field, uri zapcore.Field, status zapcore.Field) {
	zapLog.Error(message, method, uri, status)
}
