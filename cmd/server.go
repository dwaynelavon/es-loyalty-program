package main

import (
	"context"
	"fmt"
	"path"

	firebase "firebase.google.com/go"
	"github.com/dwaynelavon/es-loyalty-program/config"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty/aggregate"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

func main() {
	firebaseApp, err := newFirebaseApp()
	if err != nil {
		fmt.Printf("%v\n", err.Error())
		panic("unable to instantiate Firebase")
	}

	logger, _ := zap.NewDevelopment()

	params := aggregate.RepositoryParams{
		Store:  firebasestore.NewStore(firebaseApp),
		Logger: logger,
	}

	userRepository := aggregate.NewRepository(params)
	aggregateID, version, err := userRepository.Apply(context.TODO(), &loyalty.CreateUser{
		CommandModel: loyalty.CommandModel{
			ID: uuid.New().String(),
		},
		Username: "admin",
	})
	if err != nil {
		wrappedErr := errors.Wrap(err, "error occured while trying to apply command")
		logger.Error(wrappedErr.Error())
		return
	}
	logger.Sugar().Infof("Event saved with the version: %v", *version)

	_, version, err = userRepository.Apply(context.TODO(), &loyalty.DeleteUser{
		CommandModel: loyalty.CommandModel{
			ID: *aggregateID,
		},
	})
	if err != nil {
		wrappedErr := errors.Wrap(err, "error occured while trying to apply command")
		logger.Error(wrappedErr.Error())
		return
	}
	logger.Sugar().Infof("Event saved with the version: %v", *version)
}

func newFirebaseApp() (*firebase.App, error) {
	config.LoadEnvWithPath("../config/.env")
	configReader := config.NewReader()

	firebaseConfigFile, configErr := configReader.ReadFirebaseCredentialsFileLocation()
	if configErr != nil {
		return nil, configErr
	}

	var (
		opt  = &[]option.ClientOption{option.WithCredentialsFile(path.Join("../config", *firebaseConfigFile))}
		conf = &firebase.Config{}
		ctx  = context.Background()
	)

	app, err := firebase.NewApp(ctx, conf, *opt...)
	if err != nil {
		return nil, fmt.Errorf("Error initializing firebase app: %v", err)
	}

	return app, nil
}
