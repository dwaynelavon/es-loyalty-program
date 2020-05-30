package eventsource

import "context"

// Aggregate consumes a command and emits Events
type Aggregate interface {
	// EventVersion returns the current event version
	EventVersion() int

	// Apply is a method used to apply a history of events to an aggregate instance
	Apply(History) error
}

// Applier provides an outline for event applier behavior
type Applier interface {
	Apply(Aggregate) error
}

type CommandHandler interface {
	// Apply applies a command to an aggregate to generate a new set of events
	Handle(context.Context, Command) ([]Event, error)

	// CommandsHandled returns a list of commands the CommandHandler accepts
	CommandsHandled() []Command
}
