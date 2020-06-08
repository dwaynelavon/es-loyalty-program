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
		},
	}
}

func (h *userEventHandler) Handle(ctx context.Context, event eventsource.Event) error {
	switch event.EventType {
	case userCreatedEventType:
		p, errPayload := deserialize(event.EventType, event.Payload)
		if errPayload != nil {
			return errPayload
		}
		if p.Email == nil || p.Username == nil || p.ReferralCode == nil {
			return newInvalidPayloadError(event.EventType)
		}
		return h.readRepo.CreateUser(ctx, UserDTO{
			UserID:       event.AggregateID,
			Username:     *p.Username,
			Email:        *p.Email,
			CreatedAt:    *p.CreatedAt,
			UpdatedAt:    *p.UpdatedAt,
			ReferralCode: *p.ReferralCode,
		})
	case userReferralCompletedEventType:
		p, errPayload := deserialize(event.EventType, event.Payload)
		if errPayload != nil {
			return errPayload
		}
		if p.ReferralID == nil || *p.ReferralID == "" {
			return newInvalidPayloadError(event.EventType)
		}
		return h.readRepo.UpdateReferralStatus(ctx, event.AggregateID, *p.ReferralID, ReferralStatusCompleted)
	case userReferralCreatedEventType:
		p, errPayload := deserialize(event.EventType, event.Payload)
		if errPayload != nil {
			return errPayload
		}
		if eventsource.IsStringEmpty(p.ReferredUserEmail) ||
			eventsource.IsStringEmpty(p.ReferralCode) ||
			eventsource.IsStringEmpty(p.ReferralID) {
			return newInvalidPayloadError(event.EventType)
		}
		status, errStatus := getReferralStatus(p.ReferralStatus)
		if errStatus != nil {
			return newInvalidPayloadError(event.EventType)
		}
		return h.readRepo.CreateReferral(ctx, event.AggregateID, Referral{
			ID:                *p.ReferralID,
			ReferralCode:      *p.ReferralCode,
			ReferredUserEmail: *p.ReferredUserEmail,
			Status:            status,
			CreatedAt:         *p.CreatedAt,
			UpdatedAt:         *p.UpdatedAt,
		})
	case userDeletedEventType:
		return h.readRepo.DeleteUser(ctx, event.AggregateID)
	}

	return nil
}

func (h *userEventHandler) EventTypesHandled() []string {
	return h.eventTypesHandled
}
