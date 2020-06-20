package eventsource

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

/* ----- tests ----- */
func TestDispatch_BlankIDError(t *testing.T) {
	assert := assert.New(t)

	dispatcher := NewDispatcher(zaptest.NewLogger(t))
	err := dispatcher.Dispatch(context.Background(), &mockCommand{
		id: "",
	})

	assert.EqualError(errors.Cause(err), errBlankCommandAggID.Error())
}

func TestConnect_NoHandlerError(t *testing.T) {
	assert := assert.New(t)

	dispatcher := NewDispatcher(zaptest.NewLogger(t))
	command := &mockCommand{
		id: "123123",
	}
	err := dispatcher.Dispatch(context.Background(), command)

	assert.EqualError(
		errors.Cause(err),
		errMissingDispatchHandlerForCommand.Error(),
	)
}

func TestHandleDispatch(t *testing.T) {
	repo := new(mockRepo)
	commandHandler := newMockCommandHandler(nil)

	dispatcher := NewDispatcher(zaptest.NewLogger(t))
	dispatcher.RegisterHandler(commandHandler)

	err := dispatcher.Dispatch(
		context.Background(),
		&mockCommand{
			id: "123123",
		},
	)

	// Expect mocked functions to be called
	assert.Nil(t, err)
	repo.AssertExpectations(t)
	commandHandler.AssertExpectations(t)
}

func TestHandleDispatch_CommandHandlerError(t *testing.T) {
	assert := assert.New(t)
	repo := new(mockRepo)
	errCommand := errors.New("new command error")
	commandHandler := newMockCommandHandler(errCommand)

	dispatcher := NewDispatcher(zaptest.NewLogger(t))
	dispatcher.RegisterHandler(commandHandler)

	err := dispatcher.Dispatch(
		context.Background(),
		&mockCommand{
			id: "123123",
		},
	)

	// Expect mocked functions to be called
	assert.EqualError(errors.Cause(err), errCommand.Error())
	repo.AssertExpectations(t)
	commandHandler.AssertExpectations(t)
}

/* ----- repo ----- */
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Save(ctx context.Context, events ...Event) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockRepo) Load(ctx context.Context, aggregateID string) (Aggregate, error) {
	args := m.Called()
	return args.Get(0).(Aggregate), args.Error(1)
}

func (m *mockRepo) Apply(ctx context.Context, events ...Event) (*string, *int, error) {
	args := m.Called()
	arg0 := args.String(0)
	arg1 := args.Int(1)
	arg2 := args.Error(2)
	return &arg0, &arg1, arg2
}

/* ----- command ----- */

type mockCommand struct {
	mock.Mock
	id string
}

func (m *mockCommand) AggregateID() string {
	return m.id
}

/* ----- command handler ----- */
type mockCommandHandler struct {
	mock.Mock
}

func (m *mockCommandHandler) Handle(ctx context.Context, command Command) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockCommandHandler) CommandsHandled() []Command {
	return []Command{
		&mockCommand{
			id: "123123",
		},
	}
}

/* ----- helpers ----- */
func newMockCommandHandler(returnedError error) *mockCommandHandler {
	commandHandler := new(mockCommandHandler)
	commandHandler.On("Handle", mock.Anything, mock.Anything).Return(returnedError)
	return commandHandler
}
