package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

var (
	userDeletedEventType = "UserDeleted"
	userCreatedEventType = "UserCreated"
)

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

// Deleted events is fired when a user is deleted
type Deleted struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *Deleted) Apply(u eventsource.Aggregate) error {
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
	user.DeletedAt = p.DeletedAt
	return nil
}
