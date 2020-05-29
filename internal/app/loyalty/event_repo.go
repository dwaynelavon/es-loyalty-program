package loyalty

import (
	"context"
)

// History represents
type History []Event

// EventStore represents the method contract for interactinng with the Event store
type EventStore interface {
	Save(ctx context.Context, events ...*Event) error

	// Load retrives event records from the store and returns them in ASC order
	Load(ctx context.Context, aggregateID string) (History, error)
}
