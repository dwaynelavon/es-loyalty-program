package command

import (
	"context"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"github.com/pkg/errors"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"
)

var commandsHandled []eventsource.Command = []eventsource.Command{
	&loyalty.CompleteReferral{},
	&loyalty.CreateReferral{},
	&loyalty.CreateUser{},
	&loyalty.DeleteUser{},
	&loyalty.EarnPoints{},
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
	case *loyalty.EarnPoints:
		events, err = c.handleEarnPoints(ctx, v)
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

func (c *commandHandler) handleCompleteReferral(
	ctx context.Context,
	command *loyalty.CompleteReferral,
) ([]eventsource.Event, error) {
	aggId := command.AggregateID()
	if eventsource.IsStringEmpty(&command.ReferredUserID) {
		return nil, errors.New("ReferredUserID must be defined when completing referral")
	}
	u, err := c.loadUserAggregate(ctx, aggId)
	if err != nil {
		return nil, err
	}
	if eventsource.IsStringEmpty(u.ReferralCode) ||
		*u.ReferralCode != command.ReferredByCode {
		return nil, errors.Errorf(
			"referredBy code %v does not match user's referral code",
			&command.ReferredByCode)
	}

	a := user.NewReferralCompletedApplier(
		command.AggregateID(),
		user.UserReferralCompletedEventType,
		u.Version+1,
	)
	errSetPayload := a.SetSerializedPayload(
		getCompleteReferralPayload(
			u.Referrals,
			command.ReferredUserEmail,
			*u.ReferralCode,
		),
	)
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{a.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) handleCreateUser(
	ctx context.Context,
	command *loyalty.CreateUser,
) ([]eventsource.Event, error) {
	if eventsource.IsAnyStringEmpty(&command.Username, &command.Email) {
		return nil, errors.New("email and username must be defined when creating user")
	}

	referralCode, errReferralCode := generateReferralCode()
	if errReferralCode != nil {
		return nil, errors.New("error generating referral code")
	}
	a := user.NewCreatedApplier(
		command.AggregateID(),
		user.UserCreatedEventType,
		1,
	)
	errSetPayload := a.SetSerializedPayload(user.CreatedPayload{
		Username:       command.Username,
		Email:          command.Email,
		ReferredByCode: command.ReferredByCode,
		ReferralCode:   referralCode,
	})
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{a.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) handleCreateReferral(
	ctx context.Context,
	command *loyalty.CreateReferral,
) ([]eventsource.Event, error) {
	if eventsource.IsStringEmpty(&command.ReferredUserEmail) {
		return nil, errors.New("ReferredUserEmail must be defined when creating referral")
	}

	agg, err := c.loadUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}
	if agg.ReferralCode == nil {
		return nil, errors.Errorf(
			"user %v does not have a referral code",
			command.AggregateID())
	}

	payload := user.ReferralCreatedPayload{
		ReferralCode:      *agg.ReferralCode,
		ReferralID:        eventsource.NewUUID(),
		ReferralStatus:    string(user.ReferralStatusCreated),
		ReferredUserEmail: command.ReferredUserEmail,
	}

	a := user.NewReferralCreatedApplier(
		command.AggregateID(),
		user.UserReferralCreatedEventType,
		agg.Version+1,
	)
	errSetPayload := a.SetSerializedPayload(payload)
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{a.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) handleDeleteUser(
	ctx context.Context,
	command *loyalty.DeleteUser,
) ([]eventsource.Event, error) {
	u, err := c.loadUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	a := user.NewDeletedApplier(
		command.AggregateID(),
		user.UserDeletedEventType,
		u.Version+1,
	)
	errSetPayload := a.SetSerializedPayload(user.DeletedPayload{
		DeletedAt: time.Now(),
	})
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{a.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) handleEarnPoints(
	ctx context.Context,
	command *loyalty.EarnPoints,
) ([]eventsource.Event, error) {
	if command.Points == 0 {
		return nil, errors.New("points earned must be greater than zero")
	}

	u, err := c.loadUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	a := user.NewPointsEarnedApplier(
		command.AggregateID(),
		user.PointsEarnedEventType,
		u.Version+1,
	)
	errSetPayload := a.SetSerializedPayload(user.PointsEarnedPayload{
		PointsEarned: command.Points,
	})
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{a.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *commandHandler) loadUserAggregate(
	ctx context.Context,
	aggregateID string,
) (*user.User, error) {
	agg, err := c.repo.Load(ctx, aggregateID)
	if err != nil {
		return nil, err
	}

	user, errUser := user.AssertUserAggregate(agg)
	if errUser != nil {
		return nil, errUser
	}

	if user.DeletedAt != nil {
		return nil, errors.New("can not delete a user that's already deleted")
	}

	return user, nil
}

func (c *commandHandler) persist(
	ctx context.Context,
	events []eventsource.Event,
) error {
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

// Sometimes the user can use a referral code to sign up even if the referring user
// hasn't formally invited them. In this case we create a new referral with
// a completed status
func getCompleteReferralPayload(
	referrals []user.Referral,
	referredUserEmail string,
	referralCode string,
) interface{} {
	if referral := findReferral(referrals, referredUserEmail); referral != nil {
		return user.ReferralCompletedPayload{
			ReferralID: referral.ID,
		}
	}
	return user.ReferralCreatedPayload{
		ReferredUserEmail: referredUserEmail,
		ReferralCode:      referralCode,
		ReferralStatus:    string(user.ReferralStatusCompleted),
		ReferralID:        eventsource.NewUUID(),
	}
}

func generateReferralCode() (string, error) {
	return shortid.Generate()
}

func findReferral(
	referrals []user.Referral,
	referredUserEmail string,
) (referral *user.Referral) {
	for _, v := range referrals {
		if v.ReferredUserEmail == referredUserEmail {
			r := v
			referral = &r
		}
	}
	return
}
