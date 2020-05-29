package aggregate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/jinzhu/copier"
)

type applier interface {
	Apply(u *User) error
}

// UserEventRepo is the main abstraction for loading and saving events
type UserEventRepo interface {
	// Save persists the events into the underlying Store
	Save(ctx context.Context, events ...*loyalty.Record) error

	// Load retrieves the specified aggregate from the underlying store
	Load(ctx context.Context, aggregateID string) (*User, error)

	// Apply executes the command specified and returns the current version of the aggregate
	Apply(ctx context.Context, command loyalty.Command) (*string, *int, error)
}

// User encapsulates account information about an application user
type User struct {
	ID        string     `json:"id"`
	Version   int        `json:"version"`
	Username  string     `json:"username"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

// NewUser creates a new instance of the User aggregate
func NewUser(id string) *User {
	return &User{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

var (
	userCreatedEventType = "UserCreated"
	userDeletedEventType = "UserDeleted"
)

// UserCreated events is fired when a new user is created
type UserCreated struct {
	loyalty.Record
}

// Apply implements the applier interface
func (event *UserCreated) Apply(u *User) error {
	p := deserializePayload(event.Payload)
	if p == nil {
		return errors.New("missing payload for event")
	}
	copier.Copy(u, p)
	return nil
}

// UserDeleted events is fired when a user is deleted
type UserDeleted struct {
	loyalty.Record
}

// Apply implements the applier interface
func (event *UserDeleted) Apply(u *User) error {
	p := deserializePayload(event.Payload)
	u.DeletedAt = p.DeletedAt
	return nil
}

func deserializePayload(payload *string) *User {
	if payload == nil {
		return nil
	}
	var p User
	json.Unmarshal([]byte(*payload), &p)
	return &p
}

func getApplier(event loyalty.Record) (applier, bool) {
	switch event.EventType {
	case userCreatedEventType:
		return &UserCreated{
			Record: event,
		}, true

	case userDeletedEventType:
		return &UserDeleted{
			Record: event,
		}, true

	default:
		return nil, false
	}
}

func (u *User) BuildUserAggregate(history loyalty.History) error {

	// // Initialise an empty User on which all the events will be applied.
	// var u User

	// Every event in the slice carries a state change that has to be applied
	// on our User, in order to rebuild its current state.
	for _, h := range history {
		a, ok := getApplier(h)
		if !ok {
			return fmt.Errorf("error occured while trying to apply: %v", h.EventType)
		}
		a.Apply(u)
	}

	return nil
}

//Handle implements the CommandHandler interface
func (u *User) Handle(ctx context.Context, cmd loyalty.Command) ([]*loyalty.Record, error) {
	switch v := cmd.(type) {
	case *loyalty.CreateUser:
		userPayload := NewUser(v.AggregateID())
		userPayload.Username = v.Username
		userPayload.CreatedAt = time.Now()
		userPayload.UpdatedAt = time.Now()
		payload, errMarshal := json.Marshal(userPayload)
		if errMarshal != nil {
			return nil, fmt.Errorf("error occurred while marshaling command payload")
		}
		return []*loyalty.Record{
			loyalty.NewRecord(v.AggregateID(), userCreatedEventType, u.Version+1, payload),
		}, nil
	case *loyalty.DeleteUser:
		userPayload := NewUser(v.AggregateID())
		payload, errMarshal := json.Marshal(userPayload)
		if errMarshal != nil {
			return nil, fmt.Errorf("error occurred while marshaling command payload")
		}
		return []*loyalty.Record{
			loyalty.NewRecord(v.AggregateID(), userDeletedEventType, u.Version+1, payload),
		}, nil
	default:
		return nil, fmt.Errorf("unhandled command, %v", v)
	}
}
