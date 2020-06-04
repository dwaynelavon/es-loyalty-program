package user

import (
	"context"
	"errors"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"go.uber.org/zap"
)

type UserDTO struct {
	UserId    string    `json:"userId" firestore:"userId"`
	Username  string    `json:"username" firestore:"username"`
	Email     string    `json:"email" firestore:"email"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type ReadRepo interface {
	CreateUser(context.Context, UserDTO) error
	DeleteUser(context.Context, string) error
	Users(context.Context) ([]UserDTO, error)
}

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
		},
	}
}

func (h *userEventHandler) Handle(ctx context.Context, event eventsource.Event) error {
	switch event.EventType {
	case userCreatedEventType:
		p, errPayload := deserialize(event.Payload)
		if errPayload != nil {
			return errPayload
		}
		if p.Email == nil || p.Username == nil {
			return errors.New("invalid payload provided to userCreatedEventType")
		}
		return h.readRepo.CreateUser(ctx, UserDTO{
			UserId:    event.AggregateID,
			Username:  *p.Username,
			Email:     *p.Email,
			CreatedAt: *p.CreatedAt,
			UpdatedAt: *p.UpdatedAt,
		})
	case userDeletedEventType:
		return h.readRepo.DeleteUser(ctx, event.AggregateID)
	}

	return nil
}

func (h *userEventHandler) EventTypesHandled() []string {
	return h.eventTypesHandled
}
