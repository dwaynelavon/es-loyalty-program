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
		events, err = c.handleCreateUser(ctx, v)
	case *loyalty.DeleteUser:
		events, err = c.handleDeleteUser(ctx, v)
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

func newCreateUserPayload(username, email *string) ([]byte, error) {
	userPayload := &Payload{
		Username:  username,
		Email:     email,
		CreatedAt: eventsource.TimeNow(),
		UpdatedAt: eventsource.TimeNow(),
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: CreateUser")
	}

	return payload, nil
}

func newDeleteUserPayload() ([]byte, error) {
	userPayload := &Payload{
		DeletedAt: eventsource.TimeNow(),
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: DeleteUser")
	}

	return payload, nil
}

func (c *commandHandler) handleCreateUser(ctx context.Context, command *loyalty.CreateUser) ([]eventsource.Event, error) {
	username := command.Username
	email := command.Email
	if username == "" || email == "" {
		return nil, errors.New("email and username must be defined when creating user")
	}

	payload, errPayload := newCreateUserPayload(&command.Username, &command.Email)
	if errPayload != nil {
		return nil, errPayload
	}

	events := []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userCreatedEventType, 1, payload),
	}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) handleDeleteUser(ctx context.Context, command *loyalty.DeleteUser) ([]eventsource.Event, error) {
	payload, errPayload := newDeleteUserPayload()
	if errPayload != nil {
		return nil, errPayload
	}

	user, err := c.getUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	events := []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userDeletedEventType, user.Version+1, payload),
	}

	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) getUserAggregate(ctx context.Context, aggregateID string) (*User, error) {
	agg, err := c.repo.Load(ctx, aggregateID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error occurred while trying to handle delete user command for aggregate: %v\n",
			aggregateID,
		)
	}

	var user *User
	var ok bool
	if user, ok = agg.(*User); !ok {
		return nil, errors.New("unable to load aggregate history from the store. invalid type returned")
	}

	// Check that command is valid for the current state of the aggregate
	if user.DeletedAt != nil {
		return nil, errors.New("error executing command: can not delete a user that's already deleted")
	}

	return user, nil
}

func (c *commandHandler) persist(ctx context.Context, events []eventsource.Event) error {
	start := time.Now()
	aggregateID, version, errApply := c.repo.Apply(ctx, events...)
	if errApply != nil {
		return errApply
	}

	c.logger.Sugar().Infof(
		"saved %v event(s) for aggregate %v. current version is: %v (%v elapsed)",
		len(events),
		*aggregateID,
		*version,
		time.Since(start),
	)

	return nil
}
