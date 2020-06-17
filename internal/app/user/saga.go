package user

import (
	"context"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type saga struct {
	dispatcher        eventsource.CommandDispatcher
	repo              ReadRepo
	pointsMapping     loyalty.PointsMappingService
	logger            *zap.Logger
	eventTypesHandled []string
}

func NewSaga(
	logger *zap.Logger,
	dispatcher eventsource.CommandDispatcher,
	repo ReadRepo,
	pointsMapping loyalty.PointsMappingService,
) eventsource.EventHandler {
	return &saga{
		dispatcher:    dispatcher,
		logger:        logger,
		repo:          repo,
		pointsMapping: pointsMapping,
		eventTypesHandled: []string{
			userCreatedEventType,
		},
	}
}

func (s *saga) EventTypesHandled() []string {
	return s.eventTypesHandled
}

func (s *saga) Handle(
	ctx context.Context,
	event eventsource.Event,
) error {
	switch event.EventType {
	case userCreatedEventType:
		return s.handleUserCreatedEvent(ctx, event)
	}

	return nil
}

func (s *saga) handleUserCreatedEvent(
	ctx context.Context,
	event eventsource.Event,
) error {
	a, err := getApplier(event)
	if err != nil {
		return err
	}

	applier, ok := a.(*Created)
	if !ok {
		return errors.New("invalid applier for event provided")
	}
	p, errPayload := applier.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	if p.ReferredByCode == nil {
		return s.handleSignUpWithoutReferral(ctx, event)
	}

	referringUser, errReferringUser := s.repo.
		UserByReferralCode(ctx, *p.ReferredByCode)
	if errReferringUser != nil {
		return errors.Errorf(
			"referring user not found for referral code: %v and aggregate: %v",
			*p.ReferredByCode,
			event.AggregateID,
		)
	}

	errCompleteReferral := s.dispatcher.Dispatch(ctx, &loyalty.CompleteReferral{
		CommandModel: eventsource.CommandModel{
			ID: referringUser.UserID,
		},
		ReferredByCode:    *p.ReferredByCode,
		ReferredUserEmail: p.Email,
		ReferredUserID:    event.AggregateID,
	})
	if errCompleteReferral != nil {
		return errCompleteReferral
	}

	// Earn points for both users
	// TODO: Maybe these should be goroutines
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

	return s.dispatcher.Dispatch(ctx, &loyalty.EarnPoints{
		CommandModel: eventsource.CommandModel{
			ID: event.AggregateID,
		},
		Points: *signUpPoints,
	})
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
