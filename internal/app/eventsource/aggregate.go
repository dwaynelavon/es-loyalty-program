package eventsource

import "context"

// Aggregate consumes a command and emits Events
type Aggregate interface {
	// Apply applies a command to an aggregate to generate a new set of events
	Handle(ctx context.Context, command Command) ([]Event, error)

	// EventVersion returns the current event version
	EventVersion() int

	// BuildUserAggregate is a method used to applit a history of events to an aggregate instance
	BuildUserAggregate(history History) error
}

// Applier provides an outline for event applier behavior
type Applier interface {
	Apply(u Aggregate) error
}
