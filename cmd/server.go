package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dwaynelavon/es-loyalty-program/config"
	"github.com/dwaynelavon/es-loyalty-program/graph"
	"github.com/dwaynelavon/es-loyalty-program/graph/generated"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	firebaseEventStore "github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore/event"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore/readmodel"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

const defaultPort = "8080"

func main() {
	firebaseApp, err := newFirebaseApp()
	if err != nil {
		panic(errors.Wrap(err, "unable to instantiate Firebase"))

	}
	firestoreClient, errFirestoreClient := firebaseApp.Firestore(context.Background())
	if errFirestoreClient != nil {
		panic(errors.Wrap(errFirestoreClient, "unable to instantiate Firebase"))
	}

	logger, _ := zap.NewDevelopment()
	eventBus := connectEventBus(logger, firestoreClient)
	dispatcher := connectDispatcher(logger, firestoreClient, eventBus)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		Dispatcher: dispatcher,
	}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func connectDispatcher(logger *zap.Logger, firestoreClient *firestore.Client, eventBus eventsource.EventBus) eventsource.CommandDispatcher {
	params := loyalty.RepositoryParams{
		Store:  firebaseEventStore.NewStore(firestoreClient),
		Logger: logger,
		NewAggregate: func(id string) eventsource.Aggregate {
			return user.NewUser(id)
		},
	}
	userRepository := loyalty.NewRepository(params)
	dispatcher := eventsource.NewDispatcher(logger)
	dispatcher.RegisterHandler(user.NewUserCommandHandler(user.CommandHandlerParams{
		Repo:     userRepository,
		Logger:   logger,
		EventBus: eventBus,
	}))
	_ = dispatcher.Connect()

	return dispatcher
}

func connectEventBus(logger *zap.Logger, firestoreClient *firestore.Client) eventsource.EventBus {
	// params := loyalty.RepositoryParams{
	// 	Store:  firebasestore.NewStore(firestoreClient),
	// 	Logger: logger,
	// 	NewAggregate: func(id string) eventsource.Aggregate {
	// 		return user.NewUser(id)
	// 	},
	// }
	// userRepository := loyalty.NewRepository(params)
	bus := eventsource.NewEventBus(logger)
	bus.RegisterHandler(user.NewEventHandler(logger, readmodel.NewUserStore(firestoreClient)))
	_ = bus.Connect()

	return bus
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
