package user

import (
	"context"
	"fmt"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/pkg/errors"
)

var userDeletedEventType = "UserDeleted"

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

func handleDeleteUser(u *User, command *loyalty.DeleteUser) ([]eventsource.Event, error) {
	deletedAt := time.Now()
	userPayload := &Payload{
		DeletedAt: &deletedAt,
	}
	payload, errMarshal := serialize(userPayload)
	if errMarshal != nil {
		return nil, fmt.Errorf("error occurred while serializing command payload: DeleteUser")
	}
	return []eventsource.Event{
		*eventsource.NewEvent(command.AggregateID(), userDeletedEventType, u.Version+1, payload),
	}, nil
}

// Handle implements the Aggregate interface
func (u *User) Handle(ctx context.Context, cmd eventsource.Command) ([]eventsource.Event, error) {
	switch v := cmd.(type) {
	case *loyalty.CreateUser:
		return handleCreateUser(u, v)
	case *loyalty.DeleteUser:
		return handleDeleteUser(u, v)
	default:
		return nil, fmt.Errorf("unhandled command, %v", v)
	}
}
