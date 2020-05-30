package user

import (
	"fmt"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/pkg/errors"
)

var userCreatedEventType = "UserCreated"

// Created events is fired when a new user is created
type Created struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *Created) Apply(u eventsource.Aggregate) error {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return errors.New("invalid aggregate type being applied to a User")
	}
	p, errDeserialize := deserialize(event.Payload)
	if errDeserialize != nil {
		return errors.Wrapf(
			errDeserialize,
			"error occured while trying to deserialize payload for event type: %v",
			event.EventType,
		)
	}
	if p == nil {
		return errors.New("missing payload for event")
	}
	user.Version = event.Version
	return nil
}

func handleCreateUser(u *User, command *loyalty.CreateUser) ([]eventsource.Event, error) {
	userPayload := &Payload{
		Username:  command.Username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: CreateUser")
	}
	return []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userCreatedEventType, u.Version+1, payload),
	}, nil
}
