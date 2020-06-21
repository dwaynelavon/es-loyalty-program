package dependency

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dwaynelavon/es-loyalty-program/config"
	"github.com/dwaynelavon/es-loyalty-program/graph"
	"github.com/dwaynelavon/es-loyalty-program/graph/generated"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
)

var defaultPort = "8080"

func RegisterRoutes(
	logger *zap.Logger,
	firestoreClient *firestore.Client,
	dispatcher eventsource.CommandDispatcher,
	userReadModel user.ReadModel,
) {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Build server
	graphResolver := &graph.Resolver{
		UserReadModel: userReadModel,
		Dispatcher:    dispatcher,
	}
	generatedConfig := generated.Config{
		Resolvers: graphResolver,
	}
	schema := generated.NewExecutableSchema(generatedConfig)
	srv := handler.NewDefaultServer(schema)
	srv.SetErrorPresenter(errorPresenterWithLogger(logger))

	// Handlers
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func LoadEnv() error {
	errLoadEnv := config.LoadEnvWithPath("../config/.env")
	if errLoadEnv != nil {
		return errors.New("unable to load environment variables")
	}
	return nil
}

func errorPresenterWithLogger(
	logger *zap.Logger,
) func(ctx context.Context, err error) *gqlerror.Error {
	return func(ctx context.Context, err error) *gqlerror.Error {
		logger.Error(err.Error())
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return gqlerror.ErrorPathf(
				graphql.GetFieldContext(ctx).Path(),
				"Request timeout. Check network connection")
		}

		return graphql.DefaultErrorPresenter(ctx, err)
	}
}
