package eventsource

// Command encapsulates the data to mutate an aggregate
type Command interface {
	// AggregateID represents the id of the aggregate to apply to
	AggregateID() string
}

// CommandModel provides an embeddable struct that implements Command
type CommandModel struct {
	// ID contains the aggregate id
	ID string
}

// AggregateID implements the Command interface; returns the aggregate id
func (m CommandModel) AggregateID() string {
	return m.ID
}
