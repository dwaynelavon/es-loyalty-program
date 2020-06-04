package user

import (
	"context"
	"fmt"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var commandsHandled []eventsource.Command = []eventsource.Command{
	&loyalty.CreateUser{},
	&loyalty.DeleteUser{},
}

type commandHandler struct {
	eventBus eventsource.EventBus
	repo     eventsource.EventRepo
	logger   *zap.Logger
}

type CommandHandlerParams struct {
	EventBus eventsource.EventBus
	Repo     eventsource.EventRepo
	Logger   *zap.Logger
}

func NewUserCommandHandler(params CommandHandlerParams) eventsource.CommandHandler {
	return &commandHandler{
		eventBus: params.EventBus,
		repo:     params.Repo,
		logger:   params.Logger,
	}
}

// Handle implements the Aggregate interface. Unhandled events fall through safely
func (c *commandHandler) Handle(ctx context.Context, cmd eventsource.Command) error {
	var err error
	events := []eventsource.Event{}

	switch v := cmd.(type) {
	case *loyalty.CreateUser:
		events, err = c.handleCreateUser(ctx, c, v)
	case *loyalty.DeleteUser:
		events, err = c.handleDeleteUser(ctx, c, v)
	}

	if err != nil {
		return err
	}

	return c.eventBus.Publish(events)
}

// CommandsHandled implements the command handler interface
func (c *commandHandler) CommandsHandled() []eventsource.Command {
	return commandsHandled
}

func getApplier(event eventsource.Event) (eventsource.Applier, bool) {
	switch event.EventType {
	case userCreatedEventType:
		return &Created{
			Event: event,
		}, true

	case userDeletedEventType:
		return &Deleted{
			Event: event,
		}, true

	default:
		return nil, false
	}
}

func (c *commandHandler) handleCreateUser(ctx context.Context, handler *commandHandler, command *loyalty.CreateUser) ([]eventsource.Event, error) {
	userPayload := &Payload{
		Username:  &command.Username,
		Email:     &command.Email,
		CreatedAt: eventsource.TimeNow(),
		UpdatedAt: eventsource.TimeNow(),
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: CreateUser")
	}
	events := []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userCreatedEventType, 1, payload),
	}

	start := time.Now()
	aggregateID, version, errApply := c.repo.Apply(ctx, events...)
	if errApply != nil {
		return nil, errApply
	}

	c.logger.Sugar().Infof(
		"saved %v event(s) for aggregate %v. current version is: %v (%v elapsed)",
		len(events),
		*aggregateID,
		*version,
		time.Since(start),
	)

	return events, nil
}

func (c *commandHandler) handleDeleteUser(ctx context.Context, handler *commandHandler, command *loyalty.DeleteUser) ([]eventsource.Event, error) {
	userPayload := &Payload{
		DeletedAt: eventsource.TimeNow(),
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: DeleteUser")
	}

	// TODO: Check to make sure record isn't already deleted
	agg, err := handler.repo.Load(ctx, command.AggregateID())
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error occurred while trying to handle delete user command for aggregate: %v\n",
			command.AggregateID(),
		)
	}

	events := []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userDeletedEventType, agg.EventVersion()+1, payload),
	}

	start := time.Now()
	aggregateID, version, errApply := c.repo.Apply(ctx, events...)
	if errApply != nil {
		return nil, errApply
	}

	c.logger.Sugar().Infof(
		"saved %v event(s) for aggregate %v. current version is: %v (%v elapsed)",
		len(events),
		*aggregateID,
		*version,
		time.Since(start),
	)

	return events, nil
}
