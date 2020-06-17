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
		return handlePointsEarned(ctx, event, h.readRepo)
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

func handlePointsEarned(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	pointsEarnedEvent := PointsEarned{
		ApplierModel: eventsource.ApplierModel{
			Event: event,
		},
	}

	p, errPayload := pointsEarnedEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	return readRepo.EarnPoints(
		ctx,
		event.AggregateID,
		p.PointsEarned,
	)
}

func handleUserCreated(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	createdEvent := Created{
		ApplierModel: eventsource.ApplierModel{
			Event: event,
		},
	}

	p, errPayload := createdEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	userDTO := UserDTO{
		UserID:         createdEvent.AggregateID,
		Username:       p.Username,
		Email:          p.Email,
		CreatedAt:      createdEvent.EventAt,
		UpdatedAt:      createdEvent.EventAt,
		ReferralCode:   p.ReferralCode,
		ReferredByCode: p.ReferredByCode,
		AggregateBase: eventsource.AggregateBase{
			Version: event.Version,
		},
	}
	return readRepo.CreateUser(
		ctx,
		userDTO,
	)
}

func handleUserReferralCompleted(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	referralCompletedEvent := ReferralCompleted{
		ApplierModel: eventsource.ApplierModel{
			Event: event,
		},
	}

	p, errPayload := referralCompletedEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	return readRepo.UpdateReferralStatus(
		ctx,
		event.AggregateID,
		p.ReferralID,
		ReferralStatusCompleted,
	)
}

func handleUserReferralCreated(
	ctx context.Context,
	event eventsource.Event,
	readRepo ReadRepo,
) error {
	referralCreatedEvent := ReferralCreated{
		ApplierModel: eventsource.ApplierModel{
			Event: event,
		},
	}

	p, errPayload := referralCreatedEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	status, errStatus := getReferralStatus(&p.ReferralStatus)
	if errStatus != nil {
		return eventsource.NewInvalidPayloadError(event.EventType, p)
	}

	return readRepo.CreateReferral(
		ctx,
		event.AggregateID,
		Referral{
			ID:                p.ReferralID,
			ReferralCode:      p.ReferralCode,
			ReferredUserEmail: p.ReferredUserEmail,
			Status:            status,
			CreatedAt:         event.EventAt,
			UpdatedAt:         event.EventAt,
		},
	)
}
