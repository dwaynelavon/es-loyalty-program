package loyalty

import (
	"context"
)

// History represents
type History []Record

// EventStore represents the method contract for interactinng with the Event store
type EventStore interface {
	Save(ctx context.Context, events ...*Record) error

	// Load retrives event records from the store and returns them in ASC order
	Load(ctx context.Context, aggregateID string) (History, error)
}
