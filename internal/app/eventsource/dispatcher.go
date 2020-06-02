package eventsource

import (
	"context"

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
		if errHandler != nil {
			d.logger.Sugar().Error(errHandler)
			return
		}

		d.info(
			"handling command %T for aggregate %v",
			command,
			command.AggregateID(),
		)

		err := handler.Handle(ctx, command)
		if err != nil {
			d.logger.Sugar().Error(err)
		}
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
