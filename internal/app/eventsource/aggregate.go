package eventsource

import "github.com/pkg/errors"

// Aggregate consumes a command and emits Events
type Aggregate interface {
	// EventVersion returns the current event version
	EventVersion() int

	// Apply is a method used to apply a history of events to an aggregate instance
	Apply(History) error
}

// AggregateBase is the base model for aggregates
type AggregateBase struct {
	Version int `json:"version" firestore:"version"`
}

// Applier provides an outline for event applier behavior
type Applier interface {
	Apply(Aggregate) error
	EventModel() Event
	SetSerializedPayload(interface{}) error
	SetPayload(*string)
}

type ApplierModel struct {
	Event
}

// NewApplierModel creates a new applier model
func NewApplierModel(event Event) *ApplierModel {
	return &ApplierModel{
		Event: event,
	}
}

func (a *ApplierModel) EventModel() Event {
	return a.Event
}

func (a *ApplierModel) PayloadErr(
	operation Operation,
	payload interface{},
) error {
	return InvalidPayloadErr(
		operation,
		errors.New("missing required fields"),
		a.AggregateID,
		payload,
	)
}
