package eventsource

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type Operation string
type Error interface {
	AddOperation(operation Operation)
	FormatErrorString(error, []Operation, ...string) string
	WrapErr(msg *string)
	Error() string
}
type ErrorBase struct {
	Err        error
	Operations []Operation
}

func NewErrorBase(err error, operation Operation) ErrorBase {
	return ErrorBase{
		Err:        err,
		Operations: []Operation{operation},
	}
}

func (e *ErrorBase) AddOperation(operation Operation) {
	e.Operations = append(e.Operations, operation)
}

func (e *ErrorBase) FormatErrorString(err error, operations []Operation, data ...string) string {
	return formatErrorString(err, operations, data...)
}

func (e *ErrorBase) WrapErr(msg *string) {
	if msg == nil {
		return
	}
	e.Err = errors.Wrapf(e.Err, *msg)
}

func (e *ErrorBase) Error() string {
	return formatErrorString(
		e.Err,
		e.Operations,
	)
}

// TODO: Implement retryable for errors that can
// be retried by the event bus on failure
type Retryable interface {
	Retryable() bool
}

func IsRetryable(err error) bool {
	r, ok := errors.Cause(err).(Retryable)
	return ok && r.Retryable()
}

type NotFound interface {
	NotFound() bool
}

func IsNotFound(err error) bool {
	r, ok := errors.Cause(err).(NotFound)
	return ok && r.NotFound()
}

/* ----- command ----- */
type commandError struct {
	ErrorBase
	Command Command
}

func (c *commandError) Error() string {
	return formatErrorString(
		c.Err,
		c.Operations,
		"aggregateId", c.Command.AggregateID(),
		"command", typeOf(c.Command),
	)
}

func CommandErr(
	operation Operation,
	err error,
	msg string,
	command Command,
) error {
	wrapperErr := errors.Wrap(err, msg)
	return &commandError{
		ErrorBase: NewErrorBase(wrapperErr, operation),
		Command:   command,
	}
}

/* ----- event ----- */
type eventError struct {
	ErrorBase
	Event Event
}

func (e *eventError) Error() string {
	return formatErrorString(
		e.Err,
		e.Operations,
		"aggregateId", e.Event.AggregateID,
		"eventType", e.Event.EventType,
	)
}

func EventErr(
	operation Operation,
	err error,
	msg string,
	event Event,
) error {
	wrapperErr := errors.Wrap(err, msg)
	return &eventError{
		ErrorBase: NewErrorBase(wrapperErr, operation),
		Event:     event,
	}
}

/* ----- aggregate not found ----- */
type aggregateNotFoundError struct {
	ErrorBase
	AggregateID string
}

func (e *aggregateNotFoundError) NotFound() bool {
	return true
}

func (e *aggregateNotFoundError) Error() string {
	return formatErrorString(
		e.Err,
		e.Operations,
		"aggregateId", e.AggregateID,
	)
}

func AggregateNotFoundErr(
	operation Operation,
	aggregateID string,
) error {
	err := errors.New("aggregate not found")
	return &aggregateNotFoundError{
		ErrorBase:   NewErrorBase(err, operation),
		AggregateID: aggregateID,
	}
}

/* ----- invalid payload ----- */
type invalidPayloadErr struct {
	ErrorBase
	AggregateID string
	Payload     interface{}
}

func (e *invalidPayloadErr) Error() string {
	return formatErrorString(
		e.Err,
		e.Operations,
		"aggregateId", e.AggregateID,
		"payload", fmt.Sprintf("%v", e.Payload),
	)
}

func InvalidPayloadErr(
	operation Operation,
	err error,
	aggregateID string,
	payload interface{},
) error {
	return &invalidPayloadErr{
		ErrorBase:   NewErrorBase(err, operation),
		AggregateID: aggregateID,
		Payload:     payload,
	}
}

/* ----- helper ----- */
func wrapErr(err error, msg *string, operation Operation) error {
	if v, ok := err.(Error); ok {
		v.WrapErr(msg)
		v.AddOperation(operation)
		return v
	}

	newErr := err
	if msg != nil {
		newErr = errors.Wrap(err, *msg)
	}

	newErrBase := NewErrorBase(newErr, operation)
	return &newErrBase
}

func formatErrorString(err error, operations []Operation, data ...string) string {
	errStr := ""
	if err == nil && len(data) == 0 {
		return ""
	}
	if err != nil {
		errStr += fmt.Sprintf("%v", err)
	}

	operationStrings := []string{}
	for _, v := range operations {
		operationStrings = append(operationStrings, string(v))
	}

	dataWithOperations := append(
		[]string{
			"operations",
			fmt.Sprintf("[%v]", strings.Join(operationStrings, ", ")),
		},
		data...,
	)
	errKeyValuePairs := []string{}
	lastIndex := len(dataWithOperations) - 1
	for i := 0; i <= lastIndex; i += 2 {
		var (
			keyValuePair      = ""
			key               = dataWithOperations[i]
			value             = "nil"
			keyValueSeparator = ": "
		)

		// If we have another index remaining,
		// use it as value
		if i+1 <= lastIndex {
			value = dataWithOperations[i+1]
		}

		// open brackets for values
		if i == 0 {
			if !IsStringEmpty(&errStr) {
				// add leading space between err and data
				keyValuePair += " "
			}
			keyValuePair += "{"
		}

		keyValuePair += key + keyValueSeparator + value

		// close brackets for values
		if i+1 >= lastIndex {
			keyValuePair += "}"
		}

		errKeyValuePairs = append(errKeyValuePairs, keyValuePair)
	}

	return errStr + strings.Join(errKeyValuePairs, ", ")
}
