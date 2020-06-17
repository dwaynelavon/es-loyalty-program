package dependency

import (
	"cloud.google.com/go/firestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/firebasestore/readmodel"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	userCommand "github.com/dwaynelavon/es-loyalty-program/internal/app/user/command"
	userEvent "github.com/dwaynelavon/es-loyalty-program/internal/app/user/event"
	"go.uber.org/zap"
)

func RegisterDispatchHandlers(
	logger *zap.Logger,
	firestoreClient *firestore.Client,
	eventBus eventsource.EventBus,
	dispatcher eventsource.CommandDispatcher,
) error {
	userRepository := newUserRepository(logger, firestoreClient)
	dispatcher.RegisterHandler(
		userCommand.NewUserCommandHandler(
			userCommand.CommandHandlerParams{
				Repo:     userRepository,
				Logger:   logger,
				EventBus: eventBus,
			},
		),
	)
	return nil
}

func RegisterEventHandlers(
	logger *zap.Logger,
	firestoreClient *firestore.Client,
	eventBus eventsource.EventBus,
	dispatcher eventsource.CommandDispatcher,
	userRepo user.ReadRepo,
	pointsMappingService loyalty.PointsMappingService,
) error {
	eventBus.RegisterHandler(userEvent.NewEventHandler(logger, readmodel.NewUserStore(firestoreClient)))
	eventBus.RegisterHandler(userEvent.NewSaga(logger, dispatcher, userRepo, pointsMappingService))
	return nil
}

func NewDispatcher(logger *zap.Logger, firestoreClient *firestore.Client) eventsource.CommandDispatcher {
	return eventsource.NewDispatcher(logger)
}

func NewEventBus(logger *zap.Logger, firestoreClient *firestore.Client) eventsource.EventBus {
	return eventsource.NewEventBus(logger)
}
