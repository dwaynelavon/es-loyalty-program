package main

import (
	"context"
	"fmt"
	"path"
	"time"

	firebase "firebase.google.com/go"
	"github.com/dwaynelavon/es-loyalty-program/config"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
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

	params := loyalty.RepositoryParams{
		Store:  firebasestore.NewStore(firebaseApp),
		Logger: logger,
		NewAggregate: func(id string) eventsource.Aggregate {
			return user.NewUser(id)
		},
	}
	userRepository := loyalty.NewRepository(params)

	dispatcher := eventsource.NewDispatcher(userRepository, logger)
	dispatcher.Register(user.NewUserCommandHandler(user.CommandHandlerParams{Repo: userRepository}))
	_ = dispatcher.Connect()

	id := eventsource.NewUUID()
	dispatcher.Dispatch(context.TODO(), &loyalty.CreateUser{
		CommandModel: eventsource.CommandModel{
			ID: id,
		},
		Username: "admin",
	})

	dispatcher.Dispatch(context.TODO(), &loyalty.DeleteUser{
		CommandModel: eventsource.CommandModel{
			ID: id,
		},
	})

	time.Sleep(3 * time.Second)
}

func newFirebaseApp() (*firebase.App, error) {
	errLoadEnv := config.LoadEnvWithPath("../config/.env")
	if errLoadEnv != nil {
		panic("unable to load environment variables")
	}
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
