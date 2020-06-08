package eventsource

import (
	"context"

	"github.com/pkg/errors"
	"github.com/reactivex/rxgo/v2"
	"go.uber.org/zap"
)

var (
	errMissingEventBusHandlers = errors.New("cannot connect a event bus without any registered handlers")
	errInvalidEvent            = errors.New("only eventsource.Event may be handled by the event bus")
	errBlankAggID              = errors.New("event may not contain a blank AggregateID")
)

type EventHandler interface {
	// Responds to published events
	Handle(context.Context, Event) error

	// EventTypesHandled returns a list of events the EventHandler accepts
	EventTypesHandled() []string
}

type EventBus interface {
	Connect() error
	Publish([]Event) error
	RegisterHandler(EventHandler)
}

// TODO: Add tests for event bus
type eventBus struct {
	logger   *zap.Logger
	handlers map[string][]EventHandler
	ch       chan rxgo.Item
	obs      rxgo.Observable
}

func NewEventBus(logger *zap.Logger) EventBus {
	ch := make(chan rxgo.Item)
	return &eventBus{
		logger:   logger,
		ch:       ch,
		obs:      rxgo.FromChannel(ch, rxgo.WithPublishStrategy()),
		handlers: make(map[string][]EventHandler),
	}
}

func (e *eventBus) Publish(events []Event) error {
	var err error
	for _, v := range events {
		if v.AggregateID == "" {
			err = errBlankAggID
			continue
		}
		e.ch <- rxgo.Of(v)
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

func (e *eventBus) Connect() error {
	if len(e.handlers) == 0 {
		return errMissingEventBusHandlers
	}

	e.obs.
		Filter(filterInvalidEventWithLogger(e.logger)).
		DoOnNext(handlePublishWithBus(e))

	e.obs.Connect()
	return nil
}

func (e *eventBus) getHandlersByEvent(event Event) ([]EventHandler, error) {
	if handlers, ok := e.handlers[event.EventType]; ok {
		return handlers, nil
	}
	return nil, errors.Errorf("no handlers registered for event of type %T", e)
}

func handlePublishWithBus(e *eventBus) rxgo.NextFunc {
	return func(item interface{}) {
		var (
			event = item.(Event)
		)

		handlers, errHandler := e.getHandlersByEvent(event)
		if errHandler != nil || len(handlers) == 0 {
			e.logger.Sugar().Error(errHandler)
			return
		}

		for _, v := range handlers {
			errHandle := v.Handle(context.Background(), event)

			if errHandle != nil {
				wrappedError := errors.Wrapf(errHandle, "error occurred handling event %T for aggregate %v with handler %T",
					event,
					event.AggregateID,
					v,
				)
				e.logger.
					Sugar().Error(wrappedError)
				return
			}

			e.logger.Sugar().Infof(
				"handled event %T for aggregate %v with handler %T",
				event,
				event.AggregateID,
				v,
			)
		}
	}
}

func filterInvalidEventWithLogger(logger *zap.Logger) rxgo.Predicate {
	return func(item interface{}) bool {
		var ok bool
		var event Event
		if event, ok = item.(Event); !ok {
			logger.Error(errInvalidEvent.Error())
			return false
		}
		if event.AggregateID == "" {
			logger.Error(errBlankAggID.Error())
			return false
		}

		return true
	}
}
