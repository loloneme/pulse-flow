package logger

import (
	"os"

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
	} else {
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}

		config.InitialFields = map[string]interface{}{
			"service": "pulse-flow",
		}

		logger, err = config.Build()
	}

	if err != nil {
		return err
	}

	Log = logger
	return nil
}
