package eventsource

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/reactivex/rxgo/v2"
	"go.uber.org/zap"
)

var (
	errNoHandlersRegistered = errors.New("cannot connect a dispatcher with no handlers registered")
	errInvalidCommand       = errors.New("only eventsource.commands may be handled by the dispatcher")
	errBlankID              = errors.New("command provided to repository.Apply may not contain a blank AggregateID")
)

type CommandDispatcher interface {
	Dispatch(context.Context, Command) error
	RegisterHandler(CommandHandler)
	Connect() error
}

type dispatcher struct {
	handlers map[string]CommandHandler
	ch       chan rxgo.Item
	obs      rxgo.Observable
	logger   *zap.Logger
	repo     EventRepo
}

func NewDispatcher(repo EventRepo, logger *zap.Logger) *dispatcher {
	ch := make(chan rxgo.Item)
	return &dispatcher{
		ch:       ch,
		obs:      rxgo.FromChannel(ch, rxgo.WithPublishStrategy()),
		handlers: make(map[string]CommandHandler),
		logger:   logger,
		repo:     repo,
	}
}

type CommandDescriptor struct {
	Ctx     context.Context
	Command Command
}

/* ----- exported ----- */
func (d *dispatcher) Connect() error {
	if len(d.handlers) == 0 {
		return errNoHandlersRegistered
	}

	d.obs.
		Filter(filterInvalidCommandWithLogger(d.logger)).
		DoOnNext(handlerDispatchWithDispatcher(d))
	d.obs.Connect()

	return nil
}

func (d *dispatcher) Dispatch(ctx context.Context, cmd Command) error {
	d.ch <- rxgo.Of(CommandDescriptor{
		Ctx:     ctx,
		Command: cmd,
	})

	return nil
}

func (d *dispatcher) RegisterHandler(c CommandHandler) {
	commands := c.CommandsHandled()
	for _, v := range commands {
		typeName := typeOf(v)
		d.handlers[typeName] = c
	}
}

func (d *dispatcher) check(err error) (error, bool) {
	if err != nil {
		wrappedErr := errors.Wrap(err, "Dispatcher error:")
		d.logger.Sugar().Error(wrappedErr)
		return wrappedErr, false
	}
	return nil, true
}

/* ----- local ----- */
func (d *dispatcher) getHandler(command Command) (CommandHandler, error) {
	handler, ok := d.handlers[typeOf(command)]
	if !ok {
		return nil, errors.New("command does not have a registered command handler")
	}
	return handler, nil
}

func (d *dispatcher) info(s string, args ...interface{}) {
	d.logger.Sugar().Infof(s, args...)
}

func handlerDispatchWithDispatcher(d *dispatcher) rxgo.NextFunc {
	return func(item interface{}) {
		var (
			descriptor = item.(CommandDescriptor)
			command    = descriptor.Command
			ctx        = descriptor.Ctx
		)
		handler, errHandler := d.getHandler(command)
		if _, ok := d.check(errHandler); !ok {
			return
		}

		d.info(
			"handling command %T for aggregate %v",
			command,
			command.AggregateID(),
		)

		start := time.Now()
		events, err := handler.Handle(ctx, command)
		if _, ok := d.check(err); !ok {
			return
		}

		/*
			TODO: if there is a requirement to have more than one repository.
			this apply operation can be moved out of the dispatcher. In an effort
			to reduce some intial complexity, we can continue to handler here for now
		*/
		aggregateID, version, errApply := d.repo.Apply(ctx, events...)
		if _, ok := d.check(errApply); !ok {
			return
		}
		if aggregateID == nil || version == nil {
			return
		}

		d.info(
			"saved %v event(s) for aggregate %v. current version is: %v (%v elapsed)",
			len(events),
			*aggregateID,
			*version,
			time.Since(start),
		)
	}
}

func filterInvalidCommandWithLogger(logger *zap.Logger) rxgo.Predicate {
	return func(item interface{}) bool {
		var ok bool
		var descriptor CommandDescriptor
		if descriptor, ok = item.(CommandDescriptor); !ok {
			logger.Error(errInvalidCommand.Error())
			return false
		}
		if descriptor.Command.AggregateID() == "" {
			logger.Error(errBlankID.Error())
			return false
		}

		return true
	}
}
