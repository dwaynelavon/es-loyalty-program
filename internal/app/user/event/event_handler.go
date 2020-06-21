package event

import (
	"context"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userEventHandler struct {
	readRepo          user.ReadRepo
	logger            *zap.Logger
	sLogger           *zap.SugaredLogger
	eventStore        eventsource.EventStore
	eventTypesHandled []string
	isReadModelSynced bool
}

func NewEventHandler(
	logger *zap.Logger,
	readRepo user.ReadRepo,
	eventStore user.EventStore,
) eventsource.EventHandler {
	return &userEventHandler{
		readRepo:   readRepo,
		eventStore: eventStore,
		logger:     logger,
		sLogger:    logger.Sugar(),
		eventTypesHandled: []string{
			user.UserCreatedEventType,
			user.UserDeletedEventType,
			user.UserReferralCreatedEventType,
			user.UserReferralCompletedEventType,
			user.PointsEarnedEventType,
		},
	}
}

func (h *userEventHandler) loadAggregate(
	ctx context.Context,
	aggregateID string,
) (*user.DTO, error) {
	// TODO: cache this value since it's read multiple times during the same request
	aggregate, errAggregate := h.readRepo.User(ctx, aggregateID)
	if errAggregate != nil {
		if status.Code(errAggregate) == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(
			errAggregate,
			"unable to load aggregate from read model",
		)
	}

	return aggregate, nil
}

func (h *userEventHandler) Sync(
	ctx context.Context,
	aggregateID string,
) error {
	h.logger.Info(
		"syncing read model",
		zap.String("aggregateId", aggregateID),
	)

	aggregate, errAggregate := h.loadAggregate(ctx, aggregateID)
	if errAggregate != nil {
		return errAggregate
	}
	if aggregate == nil {
		return nil
	}

	history, errHistory := h.eventStore.Load(
		ctx,
		aggregateID,
		aggregate.Version,
	)
	if errHistory != nil {
		return errors.Wrap(errHistory, "unable to load aggregate history")
	}

	for _, v := range history {
		errHandle := h.handleEvent(ctx, v)
		if errHandle != nil {
			return errHandle
		}
	}

	h.isReadModelSynced = true
	return nil
}

func (h *userEventHandler) Handle(
	ctx context.Context,
	event eventsource.Event,
) error {
	var errSync error
	if !h.isReadModelSynced {
		errSync = h.Sync(ctx, event.AggregateID)
		if errSync != nil {
			return errors.Wrap(errSync, "unable to sync read model with event store")
		}
		return nil
	}

	return h.handleEvent(ctx, event)
}

func (h *userEventHandler) EventTypesHandled() []string {
	return h.eventTypesHandled
}

/* ----- handlers ----- */

func (h *userEventHandler) handleEvent(
	ctx context.Context,
	event eventsource.Event,
) error {
	aggregate, errAggregate := h.loadAggregate(ctx, event.AggregateID)
	if errAggregate != nil {
		return errAggregate
	}

	switch event.EventType {
	case user.PointsEarnedEventType:
		return handlePointsEarned(ctx, event, h.readRepo, aggregate)

	case user.UserCreatedEventType:
		return handleUserCreated(ctx, event, h.readRepo)

	case user.UserReferralCompletedEventType:
		return handleUserReferralCompleted(ctx, event, h.readRepo, aggregate)

	case user.UserReferralCreatedEventType:
		return handleUserReferralCreated(ctx, event, h.readRepo, aggregate)

	case user.UserDeletedEventType:
		return h.readRepo.DeleteUser(ctx, event.AggregateID)
	}

	return nil
}

func handlePointsEarned(
	ctx context.Context,
	event eventsource.Event,
	readRepo user.ReadRepo,
	aggregate *user.DTO,
) error {
	var operation eventsource.Operation = "user.handlePointsEarned"
	if aggregate == nil {
		return eventsource.AggregateNotFoundErr(operation, event.AggregateID)
	}

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
		aggregate.Version+1,
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
	aggregate *user.DTO,
) error {
	var operation eventsource.Operation = "user.handleUserReferralCompleted"

	if aggregate == nil {
		return eventsource.AggregateNotFoundErr(operation, event.AggregateID)
	}

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
		aggregate.Version+1,
	)
}

func handleUserReferralCreated(
	ctx context.Context,
	event eventsource.Event,
	readRepo user.ReadRepo,
	aggregate *user.DTO,
) error {
	var operation eventsource.Operation = "user.handlerUserReferralCreated"

	if aggregate == nil {
		return eventsource.AggregateNotFoundErr(operation, event.AggregateID)
	}

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
		return eventsource.InvalidPayloadErr(
			operation,
			errStatus,
			event.AggregateID,
			p,
		)
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
		aggregate.Version+1,
	)
}
