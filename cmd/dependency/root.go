package dependency

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var defaultAppLevel = "development"

func NewLogger(lc fx.Lifecycle) (*zap.Logger, error) {
	appLevel, appLevelExists := os.LookupEnv("APP_LEVEL")
	if !appLevelExists {
		appLevel = defaultAppLevel
	}

	var logger *zap.Logger
	var err error
	if appLevel == "production" {
		logger, err = zap.NewProduction()
	}

	if appLevel == "development" {
		logger, err = zap.NewDevelopment()
	}

	if appLevel != "development" && appLevel != "production" {
		logger = nil
		err = fmt.Errorf("Incorrect logging level provided")
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return logger.Sync()
		},
	})

	return logger, err
}
