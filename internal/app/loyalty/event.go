package loyalty

import (
	"context"
	"time"
)

// Event represents a single event applied to an aggregate
type Event interface {
	// // AggregateID returns the id of the aggregate referenced by the event
	// AggregateID string

	// // Event type descibes the type of event that occurred
	// EventType string

	// // Version contains the version number of this event
	// Version int

	// // At indicates when the event occurred
	// EventAt time.Time

	// // Data contains extra serialized data related to the specific event. Optional
	// Payload *string
}

// Event contains data related to a single event
type Event struct {
	// // AggregateID returns the id of the aggregate referenced by the event
	// A string `firestore:"aggregateId"`
	// // Event type descibes the type of event that occurred
	// T string `firestore:"eventType"`
	// // Version contains the version number of this event
	// V int `firestore:"version"`
	// // At indicates when the event occurred
	// At time.Time `firestore:"at"`
	// // Payload contains extra serialized data related to the specific event. Optional
	// P *string `firestore:"payload"`

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

// // NewEvent creates a new event model. Events are the models to be applied to an Aggregate
// func NewEvent(id string, eventType string, version int, payload []byte) *Event {
// 	payloadStr := string(payload)
// 	return &Event{
// 		AggregateID: id,
// 		EventType:   eventType,
// 		Version:     version,
// 		EventAt:     time.Now(),
// 		Payload:     &payloadStr,
// 	}
// }

// NewRecord creates a new event model. Events are the models to be applied to an Aggregate
func NewRecord(id string, eventType string, version int, payload []byte) *Event {
	payloadStr := string(payload)
	return &Event{
		AggregateID: id,
		EventType:   eventType,
		Version:     version,
		EventAt:     time.Now(),
		Payload:     &payloadStr,
	}
}

// // AggregateID implements the Event interface
// func (r Event) AggregateID() string {
// 	return r.A
// }

// // Version implements the Event interface
// func (r Event) Version() int {
// 	return r.V
// }

// // EventAt implements the Event interface
// func (r Event) EventAt() time.Time {
// 	return r.At
// }

// //EventType implements the Event interface
// func (r Event) EventType() string {
// 	return r.T
// }

// //Payload implements the Event interface
// func (r Event) Payload() *string {
// 	return r.P
// }

// CommandModel provides an embeddable struct that implements Command
type CommandModel struct {
	// ID contains the aggregate id
	ID string
}

// AggregateID implements the Command interface; returns the aggregate id
func (m CommandModel) AggregateID() string {
	return m.ID
}

// Command encapsulates the data to mutate an aggregate
type Command interface {
	// AggregateID represents the id of the aggregate to apply to
	AggregateID() string
}

// CommandHandler consumes a command and emits Events
type CommandHandler interface {
	// Apply applies a command to an aggregate to generate a new set of events
	Handle(ctx context.Context, command Command) ([]Event, error)
}

// Aggregate represents the aggregate root in the domain driven design sense.
// It represents the current state of the domain object and can be thought of
// as a left fold over events.
type Aggregate interface {
	// On will be called for each event; returns err if the event could not be
	// applied
	On(event Event) error
}
