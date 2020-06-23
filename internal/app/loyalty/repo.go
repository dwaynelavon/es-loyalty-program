package loyalty

import (
	"context"
	"sort"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type aggregateFactory func(id string) eventsource.Aggregate

// repository provides the primary abstraction to saving and loading events
type repository struct {
	store        eventsource.EventStore
	logger       *zap.Logger
	sLogger      *zap.SugaredLogger
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
		sLogger:      p.Logger.Sugar(),
		newAggregate: p.NewAggregate,
	}
}

// Load retrieves the specified aggregate from the underlying store
func (r *repository) Load(
	ctx context.Context,
	aggregateID string,
	afterVersion int,
) (eventsource.Aggregate, error) {
	return r.load(ctx, aggregateID, afterVersion)
}

// Apply executes the command and returns the current version of the aggregate
func (r *repository) Apply(
	ctx context.Context,
	events ...eventsource.Event,
) (*string, *int, error) {
	if len(events) == 0 {
		return nil, nil, errors.New("cannot apply empty event stream")
	}

	aggregateID := events[0].AggregateID
	agg, err := r.load(ctx, aggregateID, 0)
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
			"event versions must be sequential. expected %v",
			expectedVersion,
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

func (r *repository) Save(
	ctx context.Context,
	events ...eventsource.Event,
) error {
	if len(events) == 0 {
		return nil
	}

	errSave := r.store.Save(ctx, events...)
	if errSave != nil {
		return errSave
	}

	return nil
}

func (r *repository) load(
	ctx context.Context,
	aggregateID string,
	afterVersion int,
) (eventsource.Aggregate, error) {
	var operation eventsource.Operation = "loyalty.repository.load"

	history, err := r.store.Load(ctx, aggregateID, afterVersion)
	if err != nil {
		return nil, err
	}

	entryCount := len(history)
	if entryCount == 0 {
		return nil, eventsource.AggregateNotFoundErr(
			operation,
			aggregateID,
		)
	}

	r.logger.Info(
		"loaded event(s)",
		zap.Int("count", entryCount),
		zap.String("aggregateId", aggregateID),
	)

	agg := r.newAggregate(aggregateID)
	errBuildUserAgg := agg.Apply(history)
	if errBuildUserAgg != nil {
		return nil, errBuildUserAgg
	}

	return agg, nil
}
