package event

import (
	"context"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type saga struct {
	dispatcher    eventsource.CommandDispatcher
	repo          user.ReadRepo
	pointsMapping loyalty.PointsMappingService
	logger        *zap.Logger
}

func NewSaga(
	logger *zap.Logger,
	dispatcher eventsource.CommandDispatcher,
	repo user.ReadRepo,
	pointsMapping loyalty.PointsMappingService,
) eventsource.EventHandler {
	return &saga{
		dispatcher:    dispatcher,
		logger:        logger,
		repo:          repo,
		pointsMapping: pointsMapping,
	}
}

func (s *saga) EventTypesHandled() []string {
	return []string{user.UserCreatedEventType}
}

func (s *saga) Sync(ctx context.Context, aggregateID string) error {
	panic("sync for user event saga not implemented")
}

func (s *saga) Handle(
	ctx context.Context,
	event eventsource.Event,
) error {
	switch event.EventType {
	case user.UserCreatedEventType:
		return s.handleUserCreatedEvent(ctx, event)
	}

	return nil
}

func (s *saga) handleUserCreatedEvent(
	ctx context.Context,
	event eventsource.Event,
) error {
	rawApplier, err := user.GetApplier(event)
	if err != nil {
		return err
	}

	applier, ok := rawApplier.(*user.Created)
	if !ok {
		return errors.New("invalid applier for event provided")
	}

	payload, errPayload := applier.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	if payload.ReferredByCode == nil {
		return s.handleSignUpWithoutReferral(ctx, event)
	}

	referringUser, errReferringUser := s.repo.
		UserByReferralCode(ctx, *payload.ReferredByCode)
	if errReferringUser != nil {
		return errors.Errorf(
			"referring user not found for referral code: %v",
			*payload.ReferredByCode,
		)
	}

	errCompleteReferral := s.dispatcher.Dispatch(
		ctx,
		&loyalty.CompleteReferral{
			CommandModel: eventsource.CommandModel{
				ID: referringUser.UserID,
			},
			ReferredByCode:    *payload.ReferredByCode,
			ReferredUserEmail: payload.Email,
			ReferredUserID:    event.AggregateID,
		},
	)
	if errCompleteReferral != nil {
		return errCompleteReferral
	}

	// Earn points for both users
	errEarnPointsReferrer := s.handleReferUser(ctx, event, referringUser.UserID)
	if errEarnPointsReferrer != nil {
		return errEarnPointsReferrer
	}

	errEarnPointsReferee := s.handleSignUpWithReferral(ctx, event)
	if errEarnPointsReferee != nil {
		return errEarnPointsReferee
	}

	return nil
}

func (s *saga) handleSignUpWithoutReferral(
	ctx context.Context,
	event eventsource.Event,
) error {
	signUpPoints, errSignUpPoints := s.pointsMapping.
		Map(loyalty.PointsActionSignUpWithoutReferral)
	if errSignUpPoints != nil {
		return errSignUpPoints
	}

	return s.dispatcher.Dispatch(
		ctx,
		&loyalty.EarnPoints{
			CommandModel: eventsource.CommandModel{
				ID: event.AggregateID,
			},
			Points: *signUpPoints,
		},
	)
}

func (s *saga) handleReferUser(
	ctx context.Context,
	event eventsource.Event,
	referringUserID string,
) error {
	referrerPoints, errReferrerPoints := s.pointsMapping.
		Map(loyalty.PointsActionReferUser)
	if errReferrerPoints != nil {
		return errReferrerPoints
	}

	errEarnPointsReferrer := s.dispatcher.Dispatch(ctx, &loyalty.EarnPoints{
		CommandModel: eventsource.CommandModel{
			ID: referringUserID,
		},
		Points: *referrerPoints,
	})
	if errEarnPointsReferrer != nil {
		return errEarnPointsReferrer
	}

	return nil
}

func (s *saga) handleSignUpWithReferral(
	ctx context.Context,
	event eventsource.Event,
) error {
	refereePoints, errRefereePoints := s.pointsMapping.
		Map(loyalty.PointsActionSignUpWithReferral)
	if errRefereePoints != nil {
		return errRefereePoints
	}

	errEarnPointsReferee := s.dispatcher.Dispatch(ctx, &loyalty.EarnPoints{
		CommandModel: eventsource.CommandModel{
			ID: event.AggregateID,
		},
		Points: *refereePoints,
	})
	if errEarnPointsReferee != nil {
		return errEarnPointsReferee
	}

	return nil
}
