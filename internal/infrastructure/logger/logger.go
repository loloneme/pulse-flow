package logger

import (
	"os"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init() error {
	env := os.Getenv("ENV")

	var logger *zap.Logger
	var err error

	if env == "dev" {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
		if err != nil {
			return err
		}
	} else {
		encoderConfig := ecszap.NewDefaultEncoderConfig()
		core := ecszap.NewCore(encoderConfig, zapcore.AddSync(os.Stdout), zap.InfoLevel)
		logger = zap.New(core, zap.AddCaller()).With(
			zap.String("service.name", "pulse-flow"),
			zap.String("ecs.version", "8.11.0"),
		)
	}

	Log = logger
	return nil
}
