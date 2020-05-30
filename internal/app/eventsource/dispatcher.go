package eventsource

import (
	"context"

	"github.com/pkg/errors"
	"github.com/reactivex/rxgo/v2"
	"go.uber.org/zap"
)

type CommandDispatcher interface {
	Dispatch(context.Context, Command)
	Register(CommandHandler)
	Connect() error
}

type dispatcher struct {
	handlers map[string]CommandHandler
	ch       chan rxgo.Item
	obs      rxgo.Observable
	logger   *zap.Logger
	repo     EventRepo
}

func NewDispatcher(repo EventRepo, logger *zap.Logger) CommandDispatcher {
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

func (d *dispatcher) Dispatch(ctx context.Context, cmd Command) {
	d.ch <- rxgo.Of(CommandDescriptor{
		Ctx:     ctx,
		Command: cmd,
	})
}

func (d *dispatcher) Register(c CommandHandler) {
	commands := c.CommandsHandled()
	for _, v := range commands {
		typeName := typeOf(v)
		d.handlers[typeName] = c
	}
}

func getAggregateID(c Command) (*string, error) {
	aggregateID := c.AggregateID()
	if aggregateID == "" {
		return nil, errors.New("command provided to repository.Apply may not contain a blank AggregateID")
	}
	return &aggregateID, nil
}

func (d *dispatcher) check(err error) (error, bool) {
	if err != nil {
		wrappedErr := errors.Wrap(err, "Dispatcher error:")
		d.logger.Sugar().Error(wrappedErr)
		return wrappedErr, false
	}
	return nil, true
}

func (d *dispatcher) Connect() error {
	if len(d.handlers) == 0 {
		return errors.New("cannot connect a dispatcher with no handlers registered")
	}

	d.obs.DoOnNext(func(item interface{}) {
		var descriptor CommandDescriptor
		var ok bool
		if descriptor, ok = item.(CommandDescriptor); !ok {
			d.logger.Error("only eventsource.commands may be handled by the dispatcher")
			return
		}
		command := descriptor.Command
		ctx := descriptor.Ctx
		aggregateID, errBlankID := getAggregateID(command)
		if _, ok := d.check(errBlankID); !ok {
			return
		}
		d.logger.Sugar().Infof("handling command %T for aggregate %v", command, command.AggregateID())
		handler, ok := d.handlers[typeOf(command)]
		if !ok {
			_, _ = d.check(errors.New("command does not have a registered command handler"))
			return
		}
		events, err := handler.Handle(ctx, command)
		if _, ok := d.check(err); !ok {
			return
		}
		_, version, errApply := d.repo.Apply(ctx, events...)
		if _, ok := d.check(errApply); !ok {
			return
		}
		d.logger.Sugar().Infof("saved %v events for aggregate %v. current version is: %v", len(events), aggregateID, *version)
	})

	d.obs.Connect()

	return nil
}
