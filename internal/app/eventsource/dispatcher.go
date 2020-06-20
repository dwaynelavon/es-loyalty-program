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
	sLogger  *zap.SugaredLogger
}

func NewDispatcher(logger *zap.Logger) *dispatcher {
	return &dispatcher{
		handlers: make(map[string]CommandHandler),
		logger:   logger,
		sLogger:  logger.Sugar(),
	}
}

type CommandDescriptor struct {
	Ctx     context.Context
	Command Command
}

/* ----- exported ----- */
func (d *dispatcher) Dispatch(ctx context.Context, cmd Command) error {
	var operation Operation = "eventsource.dispatcher.dispatch"

	aggregateID := cmd.AggregateID()
	if IsStringEmpty(&aggregateID) {
		return CommandErr(operation, errBlankCommandAggID, "", cmd)
	}

	handler, errHandler := d.getHandler(cmd)
	if errHandler != nil {
		return CommandErr(operation, errHandler, "unable to get command handler", cmd)
	}

	err := handler.Handle(ctx, cmd)
	if err != nil {
		return wrapErr(err, nil, operation)
	}

	d.logger.Info("command handled",
		zap.String("command", typeOf(cmd)),
		zap.String("aggregateId", cmd.AggregateID()),
	)

	return nil
}

func (d *dispatcher) RegisterHandler(c CommandHandler) {
	commands := c.CommandsHandled()
	for _, v := range commands {
		typeName := typeOf(v)
		d.handlers[typeName] = c
	}
}

func (d *dispatcher) getHandler(command Command) (CommandHandler, error) {
	handler, ok := d.handlers[typeOf(command)]
	if !ok {
		return nil, errMissingDispatchHandlerForCommand
	}

	return handler, nil
}
