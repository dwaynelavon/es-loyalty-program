package user

import (
	"context"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"go.uber.org/zap"
)

type userEventHandler struct {
	readRepo          ReadRepo
	logger            *zap.Logger
	eventTypesHandled []string
}

func NewEventHandler(logger *zap.Logger, readRepo ReadRepo) eventsource.EventHandler {
	return &userEventHandler{
		readRepo: readRepo,
		logger:   logger,
		eventTypesHandled: []string{
			userCreatedEventType,
			userDeletedEventType,
			userReferralCreatedEventType,
			userReferralCompletedEventType,
			pointsEarnedEventType,
		},
	}
}

func (h *userEventHandler) Handle(ctx context.Context, event eventsource.Event) error {
	switch event.EventType {
	case pointsEarnedEventType:
		return handlerPointsEarned(ctx, event, h.readRepo)
	case userCreatedEventType:
		return handleUserCreated(ctx, event, h.readRepo)
	case userReferralCompletedEventType:
		return handleUserReferralCompleted(ctx, event, h.readRepo)
	case userReferralCreatedEventType:
		return handleUserReferralCreated(ctx, event, h.readRepo)
	case userDeletedEventType:
		return h.readRepo.DeleteUser(ctx, event.AggregateID)
	}

	return nil
}

func (h *userEventHandler) EventTypesHandled() []string {
	return h.eventTypesHandled
}

/* ----- handlers ----- */

func handlerPointsEarned(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	p, errPayload := deserialize(event)
	if errPayload != nil ||
		p == nil ||
		p.PointsEarned == nil {
		return newInvalidPayloadError(event.EventType)
	}

	return readRepo.EarnPoints(
		ctx,
		event.AggregateID,
		uint32(*p.PointsEarned),
	)
}

func handleUserCreated(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	p, errPayload := deserialize(event)
	if errPayload != nil ||
		p == nil ||
		eventsource.IsAnyStringEmpty(
			p.Email,
			p.Username,
			p.ReferralCode,
		) {
		return newInvalidPayloadError(event.EventType)
	}

	return readRepo.CreateUser(
		ctx,
		UserDTO{
			UserID:       event.AggregateID,
			Username:     *p.Username,
			Email:        *p.Email,
			CreatedAt:    *p.CreatedAt,
			UpdatedAt:    *p.UpdatedAt,
			ReferralCode: *p.ReferralCode,
			AggregateBase: eventsource.AggregateBase{
				Version: event.Version,
			},
		},
	)
}

func handleUserReferralCompleted(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	p, errPayload := deserialize(event)
	payloadInvalid := p == nil || eventsource.IsStringEmpty(p.ReferralID)
	if errPayload != nil || payloadInvalid {
		return newInvalidPayloadError(event.EventType)
	}

	return readRepo.UpdateReferralStatus(
		ctx,
		event.AggregateID,
		*p.ReferralID,
		ReferralStatusCompleted,
	)
}

func handleUserReferralCreated(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	p, errPayload := deserialize(event)
	if errPayload != nil ||
		p == nil ||
		eventsource.IsAnyStringEmpty(
			p.ReferralID,
			p.ReferredUserEmail,
			p.ReferralCode,
			p.ReferralID,
		) {
		return newInvalidPayloadError(event.EventType)
	}

	status, errStatus := getReferralStatus(p.ReferralStatus)
	if errStatus != nil {
		return newInvalidPayloadError(event.EventType)
	}

	return readRepo.CreateReferral(
		ctx,
		event.AggregateID,
		Referral{
			ID:                *p.ReferralID,
			ReferralCode:      *p.ReferralCode,
			ReferredUserEmail: *p.ReferredUserEmail,
			Status:            status,
			CreatedAt:         *p.CreatedAt,
			UpdatedAt:         *p.UpdatedAt,
		},
	)
}
