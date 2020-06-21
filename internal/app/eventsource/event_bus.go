package eventsource

import (
	"context"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dwaynelavon/es-loyalty-program/config"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	errBlankEventAggregateID            = errors.New("event may not contain a blank AggregateID")
	errNoRegisteredEventHandlersMessage = "no handlers registered for event"
)

type EventHandler interface {
	// Responds to published events
	Handle(context.Context, Event) error

	// EventTypesHandled returns a list of events the EventHandler accepts
	EventTypesHandled() []string

	// Sync ensures the read model is correctly hydrated from the event store
	Sync(ctx context.Context, aggregateID string) error
}

type EventBus interface {
	Publish([]Event) error
	RegisterHandler(EventHandler)
}

// TODO: Add tests for event bus
type eventBus struct {
	backoffConfig *config.EventBusBackoffConfig
	sLogger       *zap.SugaredLogger
	logger        *zap.Logger
	handlers      map[string][]EventHandler
}

func NewEventBus(logger *zap.Logger, configReader *config.Reader) EventBus {
	backoffConfig, err := configReader.EventBusBackoffConfig()
	if err != nil {
		backoffConfig = &config.EventBusBackoffConfig{
			MaxRetry:              3,
			MaxElapsedMillis:      500,
			InitialIntervalMillis: 300,
		}
	}
	return &eventBus{
		backoffConfig: backoffConfig,
		sLogger:       logger.Sugar(),
		logger:        logger,
		handlers:      make(map[string][]EventHandler),
	}
}

func (e *eventBus) Publish(events []Event) error {
	var op Operation = "eventsource.Publish"

	for _, event := range events {
		if event.AggregateID == "" {
			return EventErr(
				op,
				errBlankEventAggregateID,
				StringToPointer("invalid event"),
				event,
			)
		}

		err := e.handleEvent(event)
		if err != nil {
			return EventErr(
				op,
				err,
				StringToPointer("failed to handle event"),
				event,
			)
		}
	}
	return nil
}

func (e *eventBus) RegisterHandler(handler EventHandler) {
	eventTypes := handler.EventTypesHandled()
	for _, eventType := range eventTypes {
		if v, ok := e.handlers[eventType]; ok {
			e.handlers[eventType] = append(v, handler)
			return
		}
		e.handlers[eventType] = []EventHandler{
			handler,
		}
	}
}

func (e *eventBus) getHandlersByEvent(event Event) ([]EventHandler, error) {
	var op Operation = "eventsource.getHandlersByEvent"

	if handlers, ok := e.handlers[event.EventType]; ok {
		return handlers, nil
	}
	return nil, EventErr(
		op,
		errors.New(errNoRegisteredEventHandlersMessage),
		nil,
		event,
	)
}

func (e *eventBus) handleEvent(event Event) error {
	handlers, errHandler := e.getHandlersByEvent(event)
	if errHandler != nil || len(handlers) == 0 {
		return errHandler
	}

	// This needs to be buffered or else the tryHandleEvent
	// goroutines will block until the channel is read from
	errChan := make(chan error, len(handlers))

	var wg sync.WaitGroup
	for _, v := range handlers {
		wg.Add(1)
		go e.tryHandleEvent(v, event, errChan, &wg)
	}
	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return errors.Errorf("%q event processed with %d errors", event.EventType, len(errChan))
	}

	return nil
}

func (e *eventBus) newBackOff() *backoff.ExponentialBackOff {
	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = e.backoffConfig.InitialIntervalMillis
	backOff.MaxElapsedTime = e.backoffConfig.MaxElapsedMillis

	return backOff
}

func (e *eventBus) backoffOperationWithEvent(
	retryCh chan<- error,
	handler EventHandler,
	event Event,
) backoff.Operation {
	return func() error {
		errHandle := handler.Handle(context.Background(), event)
		if errHandle != nil {
			if len(retryCh) < e.backoffConfig.MaxRetry {
				retryCh <- errHandle
				return errHandle
			}
			return backoff.Permanent(errHandle)
		}

		return nil
	}
}

func (e *eventBus) tryHandleEvent(
	handler EventHandler,
	event Event,
	out chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	retryCh := make(chan error, e.backoffConfig.MaxRetry)
	operation := e.backoffOperationWithEvent(retryCh, handler, event)

	notify := func(err error, time time.Duration) {
		wrappedError := errors.Wrap(
			err,
			"retrying handler",
		)
		e.sLogger.Error(
			wrappedError,
			zap.Duration("duration", time),
			zap.Reflect("handler", handler),
		)
	}

	errBackoff := backoff.RetryNotify(operation, e.newBackOff(), notify)
	if errBackoff != nil {
		e.logger.Error(
			errBackoff.Error(),
			zap.Reflect("handler", handler),
			zap.String("eventType", event.EventType),
		)
		out <- errBackoff
		return
	}

	e.logger.Info(
		"event handled",
		zap.String("eventType", event.EventType),
		zap.String("aggregateID", event.AggregateID),
		zap.String("handler", typeOf(handler)),
	)
}
