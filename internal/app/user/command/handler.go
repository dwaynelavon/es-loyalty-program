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

type handler struct {
	eventBus eventsource.EventBus
	repo     eventsource.EventRepo
	logger   *zap.Logger
}

type CommandHandlerParams struct {
	EventBus eventsource.EventBus
	Repo     eventsource.EventRepo
	Logger   *zap.Logger
}

func NewUserCommandHandler(
	params CommandHandlerParams,
) eventsource.CommandHandler {
	return &handler{
		eventBus: params.EventBus,
		repo:     params.Repo,
		logger:   params.Logger,
	}
}

// Handle implements the CommandHandler interface
func (c *handler) Handle(
	ctx context.Context,
	cmd eventsource.Command,
) error {
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

// CommandsHandled implements the CommandHandler interface
func (c *handler) CommandsHandled() []eventsource.Command {
	return []eventsource.Command{
		&loyalty.CompleteReferral{},
		&loyalty.CreateReferral{},
		&loyalty.CreateUser{},
		&loyalty.DeleteUser{},
		&loyalty.EarnPoints{},
	}
}

func (c *handler) handleCompleteReferral(
	ctx context.Context,
	command *loyalty.CompleteReferral,
) ([]eventsource.Event, error) {
	if eventsource.IsStringEmpty(&command.ReferredUserID) {
		return nil, errors.New(
			"ReferredUserID must be defined when completing referral",
		)
	}

	aggregate, err := c.loadUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}
	if eventsource.IsStringEmpty(aggregate.ReferralCode) ||
		*aggregate.ReferralCode != command.ReferredByCode {
		return nil, errors.Errorf(
			"referredBy code %v does not match user's referral code",
			&command.ReferredByCode,
		)
	}

	applier := user.NewReferralCompletedApplier(
		aggregate.ID,
		user.UserReferralCompletedEventType,
		aggregate.Version+1,
	)
	payload := getCompleteReferralPayload(
		aggregate.Referrals,
		command.ReferredUserEmail,
		*aggregate.ReferralCode,
	)
	errSetPayload := applier.SetSerializedPayload(payload)
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{applier.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *handler) handleCreateUser(
	ctx context.Context,
	command *loyalty.CreateUser,
) ([]eventsource.Event, error) {
	if eventsource.IsAnyStringEmpty(&command.Username, &command.Email) {
		return nil, errors.New(
			"email and username must be defined when creating user",
		)
	}

	referralCode, errReferralCode := generateReferralCode()
	if errReferralCode != nil {
		return nil, errors.New("error generating referral code")
	}

	applier := user.NewCreatedApplier(
		command.AggregateID(),
		user.UserCreatedEventType,
		1,
	)
	payload := user.CreatedPayload{
		Username:       command.Username,
		Email:          command.Email,
		ReferredByCode: command.ReferredByCode,
		ReferralCode:   referralCode,
	}
	errSetPayload := applier.SetSerializedPayload(payload)
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{applier.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *handler) handleCreateReferral(
	ctx context.Context,
	command *loyalty.CreateReferral,
) ([]eventsource.Event, error) {
	if eventsource.IsStringEmpty(&command.ReferredUserEmail) {
		return nil, errors.New(
			"ReferredUserEmail must be defined when creating referral",
		)
	}

	aggregate, err := c.loadUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	if aggregate.ReferralCode == nil {
		return nil, errors.Errorf(
			"user %v does not have a referral code",
			command.AggregateID())
	}

	applier := user.NewReferralCreatedApplier(
		command.AggregateID(),
		user.UserReferralCreatedEventType,
		aggregate.Version+1,
	)
	payload := user.ReferralCreatedPayload{
		ReferralCode:      *aggregate.ReferralCode,
		ReferralID:        eventsource.NewUUID(),
		ReferralStatus:    string(user.ReferralStatusCreated),
		ReferredUserEmail: command.ReferredUserEmail,
	}
	errSetPayload := applier.SetSerializedPayload(payload)
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{applier.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *handler) handleDeleteUser(
	ctx context.Context,
	command *loyalty.DeleteUser,
) ([]eventsource.Event, error) {
	aggregate, err := c.loadUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	applier := user.NewDeletedApplier(
		command.AggregateID(),
		user.UserDeletedEventType,
		aggregate.Version+1,
	)
	payload := user.DeletedPayload{
		DeletedAt: time.Now(),
	}
	errSetPayload := applier.SetSerializedPayload(payload)
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{applier.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *handler) handleEarnPoints(
	ctx context.Context,
	command *loyalty.EarnPoints,
) ([]eventsource.Event, error) {
	if command.Points == 0 {
		return nil, errors.New("points earned must be greater than zero")
	}

	aggregate, err := c.loadUserAggregate(ctx, command.AggregateID())
	if err != nil {
		return nil, err
	}

	applier := user.NewPointsEarnedApplier(
		command.AggregateID(),
		user.PointsEarnedEventType,
		aggregate.Version+1,
	)
	errSetPayload := applier.SetSerializedPayload(user.PointsEarnedPayload{
		PointsEarned: command.Points,
	})
	if errSetPayload != nil {
		return nil, errSetPayload
	}

	events := []eventsource.Event{applier.EventModel()}
	errSave := c.persist(ctx, events)
	if errSave != nil {
		return nil, errSave
	}

	return events, nil
}

func (c *handler) loadUserAggregate(
	ctx context.Context,
	aggregateID string,
) (*user.User, error) {
	aggregate, err := c.repo.Load(ctx, aggregateID, 0)
	if err != nil {
		return nil, err
	}

	user, errUser := user.AssertUserAggregate(aggregate)
	if errUser != nil {
		return nil, errUser
	}

	if user.DeletedAt != nil {
		return nil, errors.New("user deleted")
	}

	return user, nil
}

func (c *handler) persist(
	ctx context.Context,
	events []eventsource.Event,
) error {
	start := time.Now()
	aggregateID, version, errApply := c.repo.Apply(ctx, events...)
	if errApply != nil {
		return errApply
	}

	c.logger.Info(
		"saved event(s)",
		zap.Int("count", len(events)),
		zap.String("aggregateId", *aggregateID),
		zap.Int("currentVersion", *version),
		zap.Duration("timeElapsed", time.Since(start)),
	)

	return nil
}

/* ----- helpers ----- */

/*
	Sometimes the user can use a referral code to sign up even if the referring user hasn't formally invited them. In this case we create a new referral
	with a completed status
*/
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
