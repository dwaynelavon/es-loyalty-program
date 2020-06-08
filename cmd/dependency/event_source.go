package dependency

import (
	"cloud.google.com/go/firestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore/readmodel"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"go.uber.org/zap"
)

func RegisterDispatchHandlers(
	logger *zap.Logger,
	firestoreClient *firestore.Client,
	eventBus eventsource.EventBus,
	dispatcher eventsource.CommandDispatcher,
) error {
	userRepository := newUserRepository(logger, firestoreClient)
	dispatcher.RegisterHandler(user.NewUserCommandHandler(user.CommandHandlerParams{
		Repo:     userRepository,
		Logger:   logger,
		EventBus: eventBus,
	}))
	return nil
}

func RegisterEventHandlers(
	logger *zap.Logger,
	firestoreClient *firestore.Client,
	eventBus eventsource.EventBus,
	dispatcher eventsource.CommandDispatcher,
	userRepo user.ReadRepo,
) error {
	eventBus.RegisterHandler(user.NewEventHandler(logger, readmodel.NewUserStore(firestoreClient)))
	eventBus.RegisterHandler(user.NewSaga(logger, dispatcher, userRepo))
	return eventBus.Connect()
}

func NewDispatcher(logger *zap.Logger, firestoreClient *firestore.Client) eventsource.CommandDispatcher {
	return eventsource.NewDispatcher(logger)
}

func NewEventBus(logger *zap.Logger, firestoreClient *firestore.Client) eventsource.EventBus {
	return eventsource.NewEventBus(logger)
}
