package eventsource

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	errMissingDispatchHandlerForCommand = errors.New("command does not have a registered command handler")
	errBlankCommandAggID                = errors.New("all commands must have non-blank aggregate id")
)

type CommandHandler interface {
	// Apply applies a command to an aggregate to generate a new set of events
	Handle(context.Context, Command) error

	// CommandsHandled returns a list of commands the CommandHandler accepts
	CommandsHandled() []Command
}

type CommandDispatcher interface {
	Dispatch(context.Context, Command) error
	RegisterHandler(CommandHandler)
}

type dispatcher struct {
	handlers map[string]CommandHandler
	logger   *zap.Logger
}

func NewDispatcher(logger *zap.Logger) *dispatcher {
	return &dispatcher{
		handlers: make(map[string]CommandHandler),
		logger:   logger,
	}
}

type CommandDescriptor struct {
	Ctx     context.Context
	Command Command
}

/* ----- exported ----- */
func (d *dispatcher) Dispatch(ctx context.Context, cmd Command) error {
	if cmd.AggregateID() == "" {
		return errBlankCommandAggID
	}

	handler, errHandler := d.getHandler(cmd)
	if errHandler != nil {
		return errHandler
	}

	d.info(
		"handling command %T for aggregate %v",
		cmd,
		cmd.AggregateID(),
	)

	err := handler.Handle(ctx, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (d *dispatcher) RegisterHandler(c CommandHandler) {
	commands := c.CommandsHandled()
	for _, v := range commands {
		typeName := typeOf(v)
		d.handlers[typeName] = c
	}
}

/* ----- local ----- */

func (d *dispatcher) getHandler(command Command) (CommandHandler, error) {
	handler, ok := d.handlers[typeOf(command)]
	if !ok {
		return nil, errors.Wrapf(errMissingDispatchHandlerForCommand, "%T", command)
	}
	return handler, nil
}

func (d *dispatcher) info(s string, args ...interface{}) {
	d.logger.Sugar().Infof(s, args...)
}
