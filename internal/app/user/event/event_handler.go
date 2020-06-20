package event

import (
	"context"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"go.uber.org/zap"
)

type userEventHandler struct {
	readRepo          user.ReadRepo
	logger            *zap.Logger
	eventTypesHandled []string
}

func NewEventHandler(logger *zap.Logger, readRepo user.ReadRepo) eventsource.EventHandler {
	return &userEventHandler{
		readRepo: readRepo,
		logger:   logger,
		eventTypesHandled: []string{
			user.UserCreatedEventType,
			user.UserDeletedEventType,
			user.UserReferralCreatedEventType,
			user.UserReferralCompletedEventType,
			user.PointsEarnedEventType,
		},
	}
}

func (h *userEventHandler) Handle(ctx context.Context, event eventsource.Event) error {
	switch event.EventType {
	case user.PointsEarnedEventType:
		return handlePointsEarned(ctx, event, h.readRepo)
	case user.UserCreatedEventType:
		return handleUserCreated(ctx, event, h.readRepo)
	case user.UserReferralCompletedEventType:
		return handleUserReferralCompleted(ctx, event, h.readRepo)
	case user.UserReferralCreatedEventType:
		return handleUserReferralCreated(ctx, event, h.readRepo)
	case user.UserDeletedEventType:
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
	readRepo user.ReadRepo,
) error {
	pointsEarnedEvent := user.PointsEarned{
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
	readRepo user.ReadRepo,
) error {
	createdEvent := user.Created{
		ApplierModel: eventsource.ApplierModel{
			Event: event,
		},
	}

	p, errPayload := createdEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	userDTO := user.DTO{
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
	readRepo user.ReadRepo,
) error {
	referralCompletedEvent := user.ReferralCompleted{
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
		user.ReferralStatusCompleted,
	)
}

func handleUserReferralCreated(
	ctx context.Context,
	event eventsource.Event,
	readRepo user.ReadRepo,
) error {
	var operation eventsource.Operation = "user.handlerUserReferralCreated"
	referralCreatedEvent := user.ReferralCreated{
		ApplierModel: eventsource.ApplierModel{
			Event: event,
		},
	}

	p, errPayload := referralCreatedEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	status, errStatus := user.GetReferralStatus(&p.ReferralStatus)
	if errStatus != nil {
		return eventsource.InvalidPayloadErr(operation, errStatus, event.AggregateID, p)
	}

	return readRepo.CreateReferral(
		ctx,
		event.AggregateID,
		user.Referral{
			ID:                p.ReferralID,
			ReferralCode:      p.ReferralCode,
			ReferredUserEmail: p.ReferredUserEmail,
			Status:            status,
			CreatedAt:         event.EventAt,
			UpdatedAt:         event.EventAt,
		},
	)
}
