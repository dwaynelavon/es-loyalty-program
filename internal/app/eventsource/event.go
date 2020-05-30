package eventsource

import (
	"context"
	"time"
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

// EventStore represents the method contract for interactinng with the Event store
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

	// Event type descibes the type of event that occurred
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
