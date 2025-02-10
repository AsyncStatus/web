//nolint:errcheck //it's alright here
package log

import (
	"api/config"
	"api/config/env"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func MustNewLogger(cfg *config.Config) *Logger {
	switch cfg.AppEnv {
	case env.Dev:
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()
		return &Logger{logger}
	case env.Prod:
		logger, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()
		return &Logger{logger}
	case env.Local:
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err := config.Build()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()
		return &Logger{logger}
	default:
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()
		return &Logger{logger}
	}
}
