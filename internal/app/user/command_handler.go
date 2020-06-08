package user

import (
	"context"
	"fmt"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/pkg/errors"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"
)

var commandsHandled []eventsource.Command = []eventsource.Command{
	&loyalty.CreateUser{},
	&loyalty.DeleteUser{},
	&loyalty.CreateReferral{},
	&loyalty.CompleteReferral{},
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
	case *loyalty.CompleteReferral:
		events, err = c.handleCompleteReferral(ctx, v)
	case *loyalty.CreateUser:
		events, err = c.handleCreateUser(ctx, v)
	case *loyalty.CreateReferral:
		events, err = c.handleCreateReferral(ctx, v)
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

	case userReferralCreatedEventType:
		return &ReferralCreated{
			Event: event,
		}, true

	case userDeletedEventType:
		return &Deleted{
			Event: event,
		}, true
	case userReferralCompletedEventType:
		return &ReferralCompleted{
			Event: event,
		}, true
	default:
		return nil, false
	}
}

func (c *commandHandler) handleCompleteReferral(ctx context.Context, command *loyalty.CompleteReferral) ([]eventsource.Event, error) {
	referredUserID := command.ReferredUserID
	if referredUserID == "" {
		return nil, errors.New("referred user id must be defined when completing referral")
	}

	user, err := c.getUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}
	if user.ReferralCode == nil {
		return nil, errors.Errorf(
			"user %v does not have a referral code %v",
			command.AggregateID(),
			user.ReferralCode,
		)
	}
	if *user.ReferralCode != command.ReferredByCode {
		return nil, errors.Errorf(
			"aggregate: %v. referredByCode %v for command does not match user's referral code %v",
			command.AggregateID(),
			&command.ReferredByCode,
			*user.ReferralCode,
		)
	}

	var referral *Referral
	for _, v := range user.Referrals {
		if v.ReferredUserEmail == command.ReferredUserEmail {
			r := v
			referral = &r
		}
	}

	var serializedPayload []byte
	var errPayload error
	if referral == nil {
		serializedPayload, errPayload = newCreateReferralPayload(&command.ReferredUserEmail, user.ReferralCode, true)
	} else {
		serializedPayload, errPayload = newCompleteReferralPayload(referral.ID)
	}
	if errPayload != nil {
		return nil, errPayload
	}

	events := []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userReferralCompletedEventType, user.Version+1, serializedPayload),
	}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) handleCreateUser(ctx context.Context, command *loyalty.CreateUser) ([]eventsource.Event, error) {
	username := command.Username
	email := command.Email
	if username == "" || email == "" {
		return nil, errors.New("email and username must be defined when creating user")
	}

	payload, errPayload := newCreateUserPayload(command.AggregateID(), &command.Username, &command.Email, command.ReferredByCode)
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

func (c *commandHandler) handleCreateReferral(ctx context.Context, command *loyalty.CreateReferral) ([]eventsource.Event, error) {
	referredUserEmail := command.ReferredUserEmail
	if referredUserEmail == "" {
		return nil, errors.New("referredUserEmail must be defined when creating referral")
	}

	user, err := c.getUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}
	if user.ReferralCode == nil {
		return nil, errors.Errorf("error creating referral. user %v does not have a referral code", command.AggregateID())
	}

	payload, errPayload := newCreateReferralPayload(&command.ReferredUserEmail, user.ReferralCode, false)
	if errPayload != nil {
		return nil, errPayload
	}

	events := []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userReferralCreatedEventType, user.Version+1, payload),
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
		return nil, err
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

/* ----- helpers ----- */
func newCreateUserPayload(aggregateID string, username, email, referredByCode *string) ([]byte, error) {
	referralCode, errReferralCode := shortid.Generate()
	if errReferralCode != nil {
		return nil, errors.Wrapf(
			errReferralCode,
			"error generating referral code for aggregate: %v",
			aggregateID,
		)
	}
	userPayload := &Payload{
		Username:       username,
		Email:          email,
		ReferralCode:   &referralCode,
		ReferredByCode: referredByCode,
		CreatedAt:      eventsource.TimeNow(),
		UpdatedAt:      eventsource.TimeNow(),
	}
	payload, errMarshal := serialize("CreateUser", userPayload)
	if errMarshal != nil {
		return nil, errMarshal
	}

	return payload, nil
}

func newCompleteReferralPayload(referralID string) ([]byte, error) {
	status := string(ReferralStatusCompleted)
	referralPayload := &Payload{
		ReferralID:     &referralID,
		ReferralStatus: &status,
		CreatedAt:      eventsource.TimeNow(),
		UpdatedAt:      eventsource.TimeNow(),
	}
	payload, errMarshal := serialize("CompleteReferral", referralPayload)
	if errMarshal != nil {
		return nil, errMarshal
	}

	return payload, nil
}

func newCreateReferralPayload(referredUserEmail, referralCode *string, completed bool) ([]byte, error) {
	status := string(ReferralStatusCreated)
	if completed {
		status = string(ReferralStatusCompleted)
	}

	referralID := eventsource.NewUUID()
	referralPayload := &Payload{
		ReferralID:        &referralID,
		ReferredUserEmail: referredUserEmail,
		ReferralCode:      referralCode,
		ReferralStatus:    &status,
		CreatedAt:         eventsource.TimeNow(),
		UpdatedAt:         eventsource.TimeNow(),
	}

	payload, errMarshal := serialize("CreateReferral", referralPayload)
	if errMarshal != nil {
		return nil, errMarshal
	}

	return payload, nil
}

func newDeleteUserPayload() ([]byte, error) {
	userPayload := &Payload{
		DeletedAt: eventsource.TimeNow(),
	}
	payload, errMarshal := serialize("DeleteUser", userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: DeleteUser")
	}

	return payload, nil
}
