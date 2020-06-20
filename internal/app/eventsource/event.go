package eventsource

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// History represents
type History []Event

// EventRepo is the main abstraction for loading and saving events
type EventRepo interface {
	// Save persists the events into the underlying store
	Save(ctx context.Context, events ...Event) error

	// Load retrieves the specified aggregate from the underlying store
	Load(ctx context.Context, aggregateID string) (Aggregate, error)

	// Apply executes the events specified and returns the current version of the aggregate
	Apply(ctx context.Context, events ...Event) (*string, *int, error)
}

// EventStore represents the method contract for interacting with the Event store
type EventStore interface {
	// Save persists events to the store
	Save(context.Context, ...Event) error

	// Load retrives event records from the store and returns them in ASC order
	Load(context.Context, string) (History, error)
}

// Event contains data related to a single event
type Event struct {
	// AggregateID returns the id of the aggregate referenced by the event
	AggregateID string

	// Event type describes the type of event that occurred
	EventType string

	// Version contains the version number of this event
	Version int

	// At indicates when the event occurred
	EventAt time.Time

	// Data contains extra serialized data related to the specific event. Optional
	Payload *string
}

// NewEvent creates a new event model. Events are the models to be applied to an Aggregate
func NewEvent(id, eventType string, version int, payload []byte) *Event {
	payloadStr := string(payload)
	return &Event{
		AggregateID: id,
		EventType:   eventType,
		Version:     version,
		EventAt:     time.Now(),
		Payload:     &payloadStr,
	}
}

func (event *Event) SetPayload(payload *string) {
	event.Payload = payload
}

func (event *Event) Serialize(payload interface{}) error {
	var operation Operation = "eventsource.event.Serialize"

	serializedPayload, errMarshal := json.Marshal(payload)
	if errMarshal != nil {
		return InvalidPayloadErr(
			operation,
			errors.Wrap(errMarshal, "unable to serialize payload"),
			event.AggregateID,
			event.Payload,
		)
	}

	strPayload := string(serializedPayload)
	event.SetPayload(&strPayload)
	return nil
}

func (event *Event) Deserialize(destination interface{}) error {
	var operation Operation = "eventsource.event.Deserialize"

	if event.Payload == nil {
		return InvalidPayloadErr(
			operation,
			errors.New("event missing payload"),
			event.AggregateID,
			event.Payload,
		)
	}

	err := json.Unmarshal([]byte(*event.Payload), destination)
	if err != nil {
		return InvalidPayloadErr(
			operation,
			errors.Wrap(err, "unable to deserialize payload"),
			event.AggregateID,
			event.Payload,
		)
	}

	return nil
}
