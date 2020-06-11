package eventsource

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

var event1 = "event1"

/* ----- tests ----- */
func TestEventBus_BlankIDError(t *testing.T) {
	assert := assert.New(t)

	eventBus := NewEventBus(zaptest.NewLogger(t))
	err := eventBus.Publish([]Event{
		*NewEvent("", event1, 1, nil),
	})

	assert.EqualError(err, errBlankEventAggregateID.Error())
}

func TestEventBus_NoHandlersError(t *testing.T) {
	assert := assert.New(t)

	event := *NewEvent("abc123", event1, 1, nil)
	eventBus := NewEventBus(zaptest.NewLogger(t))
	err := eventBus.Publish([]Event{
		event,
	})

	assert.EqualError(err, fmt.Sprintf(errNoRegisteredEventHandlersMessage, event))
}

func TestEventBus_HandlePublish(t *testing.T) {
	eventHandler := newMockEventHandler(nil)

	event := *NewEvent("abc123", event1, 1, nil)
	eventBus := NewEventBus(zaptest.NewLogger(t))
	eventBus.RegisterHandler(eventHandler)

	err := eventBus.Publish([]Event{
		event,
	})

	// Expect mocked functions to be called
	assert.Nil(t, err)
	eventHandler.AssertExpectations(t)
}

/* ----- event handler ----- */
type mockEventHandler struct {
	mock.Mock
}

func (m *mockEventHandler) Handle(ctx context.Context, event Event) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockEventHandler) EventTypesHandled() []string {
	return []string{
		event1,
	}
}

/* ----- helpers ----- */
func newMockEventHandler(returnedError error) *mockEventHandler {
	eventHandler := new(mockEventHandler)
	eventHandler.On("Handle", mock.Anything, mock.Anything).Return(returnedError)
	return eventHandler
}
