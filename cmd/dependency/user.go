package dependency

import (
	"cloud.google.com/go/firestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	firebaseEventStore "github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore/event"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore/readmodel"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"go.uber.org/zap"
)

func NewUserReadRepo(firestoreClient *firestore.Client) user.ReadRepo {
	return newUserReadRepo(firestoreClient)
}

func NewUserReadModel(logger *zap.Logger, readRepo user.ReadRepo) user.ReadModel {
	return user.NewReadModel(user.ReadModelParams{
		ReadRepo: readRepo,
		Logger:   logger,
	})
}

func newUserReadRepo(firestoreClient *firestore.Client) user.ReadRepo {
	return readmodel.NewUserStore(firestoreClient)
}

func newUserRepository(logger *zap.Logger, firestoreClient *firestore.Client) eventsource.EventRepo {
	params := loyalty.RepositoryParams{
		Store:  firebaseEventStore.NewStore(firestoreClient),
		Logger: logger,
		NewAggregate: func(id string) eventsource.Aggregate {
			return user.NewUser(id)
		},
	}
	return loyalty.NewRepository(params)
}
