package logging

import (
	"dev.azure.com/pomwm/pom-tech/graviflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerBuilder struct {
	*zap.Logger

	LogLevel graviflow.Config
	LogJson  graviflow.Config
	LogDev   graviflow.Config
}

func (l *LoggerBuilder) Build() graviflow.Logger {

	zapConfig := zap.NewProductionConfig()

	zapConfig.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
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

	if l.LogDev.IsSet() && l.LogDev.BoolVal() {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Encoding = "console"

	if l.LogJson.IsSet() && l.LogJson.BoolVal() {
		zapConfig.Encoding = "json"
	}

	logger, err := zapConfig.Build()
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

func (s *stdLogger) Child(name string) graviflow.Logger {
	return &stdLogger{s.logger.Named(name)}
}
