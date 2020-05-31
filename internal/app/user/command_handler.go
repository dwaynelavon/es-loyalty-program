package user

import (
	"context"
	"fmt"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/pkg/errors"
)

var commandsHandled []eventsource.Command = []eventsource.Command{
	&loyalty.CreateUser{},
	&loyalty.DeleteUser{},
}

type EventRepo interface {
	// Save persists the events into the underlying Store
	Save(context.Context, ...eventsource.Event) error

	// Load retrieves the specified aggregate from the underlying store
	Load(context.Context, string) (eventsource.Aggregate, error)
}

type commandHandler struct {
	repo EventRepo
}

type CommandHandlerParams struct {
	Repo EventRepo
}

func NewUserCommandHandler(params CommandHandlerParams) eventsource.CommandHandler {
	return &commandHandler{
		repo: params.Repo,
	}
}

// Handle implements the Aggregate interface. Unhandled events fall through safely
func (c *commandHandler) Handle(ctx context.Context, cmd eventsource.Command) ([]eventsource.Event, error) {
	switch v := cmd.(type) {
	case *loyalty.CreateUser:
		return handleCreateUser(ctx, c, v)
	case *loyalty.DeleteUser:
		return handleDeleteUser(ctx, c, v)
	}
	return nil, nil
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

func handleCreateUser(ctx context.Context, handler *commandHandler, command *loyalty.CreateUser) ([]eventsource.Event, error) {
	userPayload := &Payload{
		Username:  &command.Username,
		CreatedAt: eventsource.TimeNow(),
		UpdatedAt: eventsource.TimeNow(),
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: CreateUser")
	}
	return []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userCreatedEventType, 1, payload),
	}, nil
}

func handleDeleteUser(ctx context.Context, handler *commandHandler, command *loyalty.DeleteUser) ([]eventsource.Event, error) {
	userPayload := &Payload{
		DeletedAt: eventsource.TimeNow(),
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: DeleteUser")
	}
	agg, err := handler.repo.Load(ctx, command.AggregateID())
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error occured while trying to handle delete user command for aggregate: %v\n",
			command.AggregateID(),
		)
	}
	return []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userDeletedEventType, agg.EventVersion()+1, payload),
	}, nil
}
