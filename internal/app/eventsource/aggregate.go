package eventsource

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
func NewApplierModel(id, eventType string, version int, payload []byte) *ApplierModel {
	return &ApplierModel{
		Event: *NewEvent(id, eventType, version, payload),
	}
}

func (a *ApplierModel) EventModel() Event {
	return a.Event
}
