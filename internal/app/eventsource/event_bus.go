package eventsource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	errBlankAggID = errors.New("event may not contain a blank AggregateID")

	backoffInitialInterval = 100 * time.Millisecond
	maxElapsedTime         = 500 * time.Millisecond
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
	logger   *zap.Logger
	handlers map[string][]EventHandler
}

func NewEventBus(logger *zap.Logger) EventBus {
	return &eventBus{
		logger:   logger,
		handlers: make(map[string][]EventHandler),
	}
}

func (e *eventBus) Publish(events []Event) error {
	var err error
	for _, event := range events {
		if event.AggregateID == "" {
			err = errBlankAggID
			continue
		}
		handlers, errHandler := e.getHandlersByEvent(event)
		if errHandler != nil || len(handlers) == 0 {
			return errHandler
		}

		var wg sync.WaitGroup
		errChan := make(chan error)
		for _, v := range handlers {
			wg.Add(1)
			go tryHandleEvent(e.logger, v, event, errChan, &wg)
		}
		wg.Wait()

		// TODO: Do something with errors
		close(errChan)
	}
	return err
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
	return nil, errors.Errorf("no handlers registered for event of type %T", e)
}

func tryHandleEvent(logger *zap.Logger, handler EventHandler, event Event, out chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = backoffInitialInterval
	backOff.MaxElapsedTime = maxElapsedTime

	operation := func() error {
		errHandle := handler.Handle(context.Background(), event)

		if errHandle != nil {
			msg := fmt.Sprintf("error occurred handling event %T with handler %T",
				event,
				handler,
			)
			logger.Error(msg, zap.Error(errHandle))
			out <- errHandle
			return errHandle
		}

		logger.Sugar().Infof(
			"handled event %T for aggregate %v with handler %T",
			event,
			event.AggregateID,
			handler,
		)

		return nil
	}

	notify := func(err error, time time.Duration) {
		logger.Sugar().Warn(fmt.Sprintf(
			"retrying handler %T for event type %v with duration %v",
			handler,
			event.EventType,
			time,
		))
	}

	errBackoff := backoff.RetryNotify(operation, backOff, notify)
	if errBackoff != nil {
		out <- errBackoff
	}
}
