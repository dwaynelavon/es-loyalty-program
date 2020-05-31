package eventsource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

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

func getMockDispatcher(t *testing.T) *dispatcher {
	var repo EventRepo = &mockRepo{}
	return NewDispatcher(repo, zaptest.NewLogger(t))
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

func (m *mockCommandHandler) Handle(ctx context.Context, command Command) ([]Event, error) {
	args := m.Called()
	return args.Get(0).([]Event), args.Error(1)
}

func (m *mockCommandHandler) CommandsHandled() []Command {
	return []Command{
		&mockCommand{
			id: "123123",
		},
	}
}

func TestConnect_NoHandlerError(t *testing.T) {
	assert := assert.New(t)

	dispatcher := getMockDispatcher(t)
	err := dispatcher.Connect()

	assert.EqualError(err, errNoHandlersRegistered.Error())

	dispatcher.RegisterHandler(&mockCommandHandler{})
	err = dispatcher.Connect()

	assert.Nil(err)
}

func TestFilterInvalidCommands_InvalidCommandError(t *testing.T) {
	assert := assert.New(t)

	filter := filterInvalidCommandWithLogger(zaptest.NewLogger(t))
	ok := filter(nil)

	assert.False(ok)
}

func TestFilterInvalidCommands_BlankIDError(t *testing.T) {
	assert := assert.New(t)

	filter := filterInvalidCommandWithLogger(zaptest.NewLogger(t))

	ok := filter(&CommandDescriptor{
		Ctx:     context.TODO(),
		Command: &mockCommand{},
	})

	assert.False(ok)
}

func TestHandleDispatch_HappyPath(t *testing.T) {

	var (
		ctx     = context.TODO()
		id      = "123123"
		version = 1
	)

	// Set up the mocked functions
	events := []Event{{
		AggregateID: id,
		Version:     version,
	}}
	repo := new(mockRepo)
	commandHandler := new(mockCommandHandler)
	commandHandler.On("Handle", mock.Anything, mock.Anything).Return(events, nil)
	repo.On("Apply", mock.Anything, mock.Anything).Return(id, version, nil)

	// Create dispatcher and register handler
	dispatcher := NewDispatcher(repo, zaptest.NewLogger(t))
	dispatcher.RegisterHandler(commandHandler)
	_ = dispatcher.Connect()

	// Create dispatch handler and dispach a command
	dispatchHandler := handlerDispatchWithDispatcher(dispatcher)
	dispatchHandler(CommandDescriptor{
		Ctx: ctx,
		Command: &mockCommand{
			id: id,
		},
	})

	// Expect mocked functions to be called
	repo.AssertExpectations(t)
	commandHandler.AssertExpectations(t)
}
