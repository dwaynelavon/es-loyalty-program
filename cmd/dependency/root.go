package dependency

import (
	"context"
	"errors"
	"os"

	"github.com/dwaynelavon/es-loyalty-program/config"
	"github.com/mattn/go-colorable"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultAppLevel = "development"

func NewLogger(lc fx.Lifecycle) (*zap.Logger, error) {
	appLevel, appLevelExists := os.LookupEnv("APP_LEVEL")
	if !appLevelExists {
		appLevel = defaultAppLevel
	}

	var config zapcore.EncoderConfig
	switch appLevel {
	case "production":
		config = zap.NewProductionEncoderConfig()
	case "development":
		config = zap.NewDevelopmentEncoderConfig()
	default:
		return nil, errors.New("incorrect log level provided")
	}

	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return logger.Sync()
		},
	})

	return logger, nil
}

func NewConfigReader() *config.Reader {
	return config.NewReader()
}
