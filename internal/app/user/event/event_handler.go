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
	isReadModelSynced bool
}

// NewEventHandler creates an instance of EventHandler
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
	}
}

// EventTypesHandled implements the EventHandler interface
func (h *userEventHandler) EventTypesHandled() []string {
	return []string{
		user.UserCreatedEventType,
		user.UserDeletedEventType,
		user.UserReferralCreatedEventType,
		user.UserReferralCompletedEventType,
		user.PointsEarnedEventType,
	}
}

// Sync implements the EventHandler interface
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

// Handle implements the EventHandler interface
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

func (h *userEventHandler) loadAggregate(
	ctx context.Context,
	aggregateID string,
) (*user.DTO, error) {
	// TODO: cache this call per request
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
		ApplierModel: *eventsource.NewApplierModel(event),
	}

	payload, errPayload := pointsEarnedEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	return readRepo.EarnPoints(
		ctx,
		event.AggregateID,
		payload.PointsEarned,
		aggregate.Version+1,
	)
}

func handleUserCreated(
	ctx context.Context,
	event eventsource.Event,
	readRepo user.ReadRepo,
) error {
	createdEvent := user.Created{
		ApplierModel: *eventsource.NewApplierModel(event),
	}

	payload, errPayload := createdEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	userDTO := user.DTO{
		UserID:         createdEvent.AggregateID,
		Username:       payload.Username,
		Email:          payload.Email,
		CreatedAt:      createdEvent.EventAt,
		UpdatedAt:      createdEvent.EventAt,
		ReferralCode:   payload.ReferralCode,
		ReferredByCode: payload.ReferredByCode,
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
		ApplierModel: *eventsource.NewApplierModel(event),
	}

	payload, errPayload := referralCompletedEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	return readRepo.UpdateReferralStatus(
		ctx,
		event.AggregateID,
		payload.ReferralID,
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
		ApplierModel: *eventsource.NewApplierModel(event),
	}

	payload, errPayload := referralCreatedEvent.GetDeserializedPayload()
	if errPayload != nil {
		return errPayload
	}

	status, errStatus := user.GetReferralStatus(&payload.ReferralStatus)
	if errStatus != nil {
		return eventsource.InvalidPayloadErr(
			operation,
			errStatus,
			event.AggregateID,
			payload,
		)
	}

	return readRepo.CreateReferral(
		ctx,
		event.AggregateID,
		user.Referral{
			ID:                payload.ReferralID,
			ReferralCode:      payload.ReferralCode,
			ReferredUserEmail: payload.ReferredUserEmail,
			Status:            status,
			CreatedAt:         event.EventAt,
			UpdatedAt:         event.EventAt,
		},
		aggregate.Version+1,
	)
}
