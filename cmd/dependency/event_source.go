package dependency

import (
	"cloud.google.com/go/firestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore/readmodel"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"go.uber.org/zap"
)

func NewDispatcher(logger *zap.Logger, firestoreClient *firestore.Client, eventBus eventsource.EventBus) eventsource.CommandDispatcher {
	userRepository := newUserRepository(logger, firestoreClient)

	dispatcher := eventsource.NewDispatcher(logger)
	dispatcher.RegisterHandler(user.NewUserCommandHandler(user.CommandHandlerParams{
		Repo:     userRepository,
		Logger:   logger,
		EventBus: eventBus,
	}))
	_ = dispatcher.Connect()

	return dispatcher
}

func NewEventBus(logger *zap.Logger, firestoreClient *firestore.Client) eventsource.EventBus {
	bus := eventsource.NewEventBus(logger)
	bus.RegisterHandler(user.NewEventHandler(logger, readmodel.NewUserStore(firestoreClient)))
	_ = bus.Connect()

	return bus
}
