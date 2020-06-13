package eventsource

import (
	"context"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	errBlankEventAggregateID            = errors.New("event may not contain a blank AggregateID")
	errNoRegisteredEventHandlersMessage = "no handlers registered for event of type %T"

	backoffInitialInterval = 100 * time.Millisecond
	maxElapsedTime         = 500 * time.Millisecond
	maxRetry               = 3
)

type EventHandler interface {
	// Responds to published events
	Handle(context.Context, Event) error

	// EventTypesHandled returns a list of events the EventHandler accepts
	EventTypesHandled() []string
}

type EventBus interface {
	Publish([]Event) error
	RegisterHandler(EventHandler)
}

// TODO: Add tests for event bus
type eventBus struct {
	sLogger  *zap.SugaredLogger
	logger   *zap.Logger
	handlers map[string][]EventHandler
}

func NewEventBus(logger *zap.Logger) EventBus {
	return &eventBus{
		sLogger:  logger.Sugar(),
		logger:   logger,
		handlers: make(map[string][]EventHandler),
	}
}

func (e *eventBus) Publish(events []Event) error {
	for _, event := range events {
		if event.AggregateID == "" {
			return errBlankEventAggregateID
		}
		err := e.handleEvent(event)
		if err != nil {
			return err
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
	if handlers, ok := e.handlers[event.EventType]; ok {
		return handlers, nil
	}
	return nil, errors.Errorf(errNoRegisteredEventHandlersMessage, event)
}

func (e *eventBus) handleEvent(event Event) error {
	handlers, errHandler := e.getHandlersByEvent(event)
	if errHandler != nil || len(handlers) == 0 {
		return errHandler
	}

	// This should be buffered or else the tryHandleEvent
	// goroutines will block until the channel is read from
	errChan := make(chan error, len(handlers))

	var wg sync.WaitGroup
	for _, v := range handlers {
		wg.Add(1)
		go tryHandleEvent(e.sLogger, v, event, errChan, &wg)
	}
	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return errors.Errorf("%q event processed with %d errors", event.EventType, len(errChan))
	}

	return nil
}

func newBackOff() *backoff.ExponentialBackOff {
	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = backoffInitialInterval
	backOff.MaxElapsedTime = maxElapsedTime

	return backOff
}

func newBackOffOperation(
	sLogger *zap.SugaredLogger, retryCh chan<- error, handler EventHandler, event Event) backoff.Operation {
	return func() error {
		errHandle := handler.Handle(context.Background(), event)
		shouldContinue := len(retryCh) < maxRetry
		if errHandle != nil {
			if shouldContinue {
				retryCh <- errHandle
				return errHandle
			}
			return backoff.Permanent(errHandle)
		}

		return nil
	}
}
func tryHandleEvent(sLogger *zap.SugaredLogger, handler EventHandler, event Event, out chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	retryCh := make(chan error, maxRetry)
	operation := newBackOffOperation(sLogger, retryCh, handler, event)

	notify := func(err error, time time.Duration) {
		wrappedError := errors.Wrapf(err,
			"retrying handler %T (%v elapsed)",
			handler,
			time)
		sLogger.Error(wrappedError)
	}

	errBackoff := backoff.RetryNotify(operation, newBackOff(), notify)
	if errBackoff != nil {
		out <- errBackoff
		return
	}

	sLogger.Infof(
		"handled event %T for aggregate %v with handler %T",
		event,
		event.AggregateID,
		handler,
	)
}
