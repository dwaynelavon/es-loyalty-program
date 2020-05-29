package aggregate

import (
	"context"
	"sort"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

/*
AggregateFactory is a function that can be used
to create a new instance of a particular domain aggregate
*/
// type aggregateFactory func() loyalty.Aggregate

// repository provides the primary abstraction to saving and loading events
type repository struct {
	store  loyalty.EventStore
	logger *zap.Logger
	// newAggregate aggregateFactory
}

// RepositoryParams represent the params are needed to instantiate a new repository
type RepositoryParams struct {
	Store  loyalty.EventStore
	Logger *zap.Logger
	// NewAggregate aggregateFactory
}

// NewRepository creates a new instance of an EventRepo
func NewRepository(p RepositoryParams) UserEventRepo {
	return &repository{
		store:  p.Store,
		logger: p.Logger,
		// newAggregate: p.NewAggregate,
	}
}

// Load retrieves the specified aggregate from the underlying store
func (r *repository) Load(ctx context.Context, aggregateID string) (*User, error) {
	agg, err := r.load(ctx, aggregateID)
	return agg, err
}

// Apply executes the command specified and returns the current version of the aggregate
func (r *repository) Apply(ctx context.Context, command loyalty.Command) (*string, *int, error) {
	if command == nil {
		return nil, nil, errors.New("command provided to repository.Apply may not be nil")
	}
	aggregateID := command.AggregateID()
	if aggregateID == "" {
		return nil, nil, errors.New("command provided to repository.Apply may not contain a blank AggregateID")
	}

	agg, err := r.load(ctx, aggregateID)
	if err != nil {
		if !loyalty.IsNotFound(err) {
			return nil, nil, err
		}
		agg = &User{}
	}

	events, err := agg.Handle(ctx, command)
	if err != nil {
		return nil, nil, err
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Version < events[j].Version
	})
	expectedVersion := agg.Version + 1
	if events[0].Version != expectedVersion {
		return nil, nil, errors.Errorf(
			"events must be sequential. event with a version: %v can not be saved after an event with version: %v",
			events[0].Version,
			agg.Version,
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

func (r *repository) Save(ctx context.Context, events ...*loyalty.Record) error {
	if len(events) == 0 {
		return nil
	}

	errSave := r.store.Save(ctx, events...)
	if errSave != nil {
		return errors.Wrapf(errSave, "error occured while trying to save %v events for aggregate: %v", len(events), events[0].AggregateID)
	}

	return nil
}

// load loads the specified aggregate from the store and returns both the Aggregate and the
// current version number of the aggregate
func (r *repository) load(ctx context.Context, aggregateID string) (*User, error) {
	history, err := r.store.Load(ctx, aggregateID)
	if err != nil {
		return nil, err
	}

	entryCount := len(history)
	if entryCount == 0 {
		return nil, loyalty.NewError(nil, loyalty.ErrAggregateNotFound, "unable to load %v", aggregateID)
	}

	r.logger.Sugar().Infof("Loaded %v event(s) for aggregate id, %v", entryCount, aggregateID)

	var user User
	errBuildUserAgg := user.BuildUserAggregate(history)
	if errBuildUserAgg != nil {
		return nil, errors.Wrapf(
			errBuildUserAgg,
			"error occured while trying to build user aggregate for aggregate: %v",
			aggregateID,
		)
	}

	return &user, nil
}
