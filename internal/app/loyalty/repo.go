package loyalty

import (
	"context"
	"sort"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

/*
AggregateFactory is a function that can be used
to create a new instance of a particular domain aggregate
*/
type aggregateFactory func(id string) eventsource.Aggregate

// repository provides the primary abstraction to saving and loading events
type repository struct {
	store        eventsource.EventStore
	logger       *zap.Logger
	newAggregate aggregateFactory
}

// RepositoryParams represent the params are needed to instantiate a new repository
type RepositoryParams struct {
	Store        eventsource.EventStore
	Logger       *zap.Logger
	NewAggregate aggregateFactory
}

// NewRepository creates a new instance of an EventRepo
func NewRepository(p RepositoryParams) eventsource.EventRepo {
	return &repository{
		store:        p.Store,
		logger:       p.Logger,
		newAggregate: p.NewAggregate,
	}
}

// Load retrieves the specified aggregate from the underlying store
func (r *repository) Load(ctx context.Context, aggregateID string) (eventsource.Aggregate, error) {
	agg, err := r.load(ctx, aggregateID)
	return agg, err
}

// Apply executes the command specified and returns the current version of the aggregate
func (r *repository) Apply(ctx context.Context, events ...eventsource.Event) (*string, *int, error) {
	if len(events) == 0 {
		return nil, nil, errors.New("cannot apply empty event stream")
	}
	aggregateID := events[0].AggregateID
	agg, err := r.load(ctx, aggregateID)
	if err != nil {
		if !eventsource.IsNotFound(err) {
			return nil, nil, err
		}
		agg = r.newAggregate(aggregateID)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Version < events[j].Version
	})
	expectedVersion := agg.EventVersion() + 1
	if events[0].Version != expectedVersion {
		return nil, nil, errors.Errorf(
			"events must be sequential. event with a version: %v can not be saved after an event with version: %v",
			events[0].Version,
			agg.EventVersion(),
		)
	}

	err = r.Save(ctx, events...)
	if err != nil {
		return nil, nil, err
	}

	if v := len(events); v > 0 {
		aggregateID = events[v-1].AggregateID
	}

	return &aggregateID, &expectedVersion, nil
}

func (r *repository) Save(ctx context.Context, events ...eventsource.Event) error {
	if len(events) == 0 {
		return nil
	}

	errSave := r.store.Save(ctx, events...)
	if errSave != nil {
		return errors.Wrapf(errSave, "error occurred while trying to save %v events for aggregate: %v", len(events), events[0].AggregateID)
	}

	return nil
}

// load loads the specified aggregate from the store and returns both the Aggregate and the
// current version number of the aggregate
func (r *repository) load(ctx context.Context, aggregateID string) (eventsource.Aggregate, error) {
	history, err := r.store.Load(ctx, aggregateID)
	if err != nil {
		return nil, err
	}

	entryCount := len(history)
	if entryCount == 0 {
		return nil, eventsource.NewError(nil, eventsource.ErrAggregateNotFound, "unable to load %v", aggregateID)
	}

	r.logger.Sugar().Infof("loaded %v event(s) for aggregate id, %v", entryCount, aggregateID)

	agg := r.newAggregate(aggregateID)
	errBuildUserAgg := agg.Apply(history)
	if errBuildUserAgg != nil {
		return nil, errors.Wrapf(
			errBuildUserAgg,
			"error occurred while trying to build user aggregate for aggregate: %v",
			aggregateID,
		)
	}

	return agg, nil
}
