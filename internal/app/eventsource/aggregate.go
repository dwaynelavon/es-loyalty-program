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
}
