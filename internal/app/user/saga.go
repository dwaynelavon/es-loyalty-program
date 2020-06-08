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
	logger            *zap.Logger
	eventTypesHandled []string
}

func NewSaga(logger *zap.Logger, dispatcher eventsource.CommandDispatcher, repo ReadRepo) eventsource.EventHandler {
	return &saga{
		dispatcher: dispatcher,
		logger:     logger,
		repo:       repo,
		eventTypesHandled: []string{
			userCreatedEventType,
		},
	}
}

func (s *saga) Handle(ctx context.Context, event eventsource.Event) error {
	switch event.EventType {
	case userCreatedEventType:
		p, errPayload := deserialize(event.EventType, event.Payload)
		if errPayload != nil {
			return errPayload
		}
		if p.ReferredByCode == nil {
			return nil
		}
		if p.Email == nil {
			return newInvalidPayloadError(event.EventType)
		}
		referringUser, errReferringUser := s.repo.UserByReferralCode(ctx, *p.ReferredByCode)
		if errReferringUser != nil {
			return errors.Errorf(
				"referring user not found for referral code: %v and aggregate: %v",
				*p.ReferredByCode,
				event.AggregateID,
			)
		}
		return s.dispatcher.Dispatch(ctx, &loyalty.CompleteReferral{
			CommandModel: eventsource.CommandModel{
				ID: referringUser.UserID,
			},
			ReferredByCode:    *p.ReferredByCode,
			ReferredUserEmail: *p.Email,
			ReferredUserID:    event.AggregateID,
		})
	}

	return nil
}

func (s *saga) EventTypesHandled() []string {
	return s.eventTypesHandled
}
