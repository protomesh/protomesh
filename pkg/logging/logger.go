package logging

import (
	"github.com/protomesh/protomesh"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerBuilder struct {
	*zap.Logger

	LogLevel protomesh.Config
	LogJson  protomesh.Config
	LogDev   protomesh.Config
}

func (l *LoggerBuilder) Build() protomesh.Logger {

	zapConfig := zap.NewProductionConfig()

	if l.LogDev.IsSet() && l.LogDev.BoolVal() {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	}

	zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	if l.LogLevel.IsSet() {
		switch l.LogLevel.StringVal() {

		case "error":
			zapConfig.Level.SetLevel(zap.ErrorLevel)

		case "info":
			zapConfig.Level.SetLevel(zap.InfoLevel)

		default:
			zapConfig.Level.SetLevel(zap.DebugLevel)

		}
	}

	zapConfig.Encoding = "console"

	if l.LogJson.IsSet() && l.LogJson.BoolVal() {
		zapConfig.Encoding = "json"
	}

	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		panic(err)
	}

	l.Logger = logger

	return &stdLogger{logger.Sugar()}

}

type stdLogger struct {
	logger *zap.SugaredLogger
}

func (s *stdLogger) Debug(message string, kv ...interface{}) {
	s.logger.Debugw(message, kv...)
}

func (s *stdLogger) Info(message string, kv ...interface{}) {
	s.logger.Infow(message, kv...)
}

func (s *stdLogger) Warn(message string, kv ...interface{}) {
	s.logger.Warnw(message, kv...)
}

func (s *stdLogger) Error(message string, kv ...interface{}) {
	s.logger.Errorw(message, kv...)
}

func (s *stdLogger) Panic(message string, kv ...interface{}) {
	s.logger.Panicw(message, kv...)
}

func (s *stdLogger) With(kv ...interface{}) protomesh.Logger {
	return &stdLogger{s.logger.With(kv...)}
}
