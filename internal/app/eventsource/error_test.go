package eventsource

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFormatErrorString_HasData(t *testing.T) {
	assert := assert.New(t)

	operations := []Operation{"eventsource.error_test"}
	err := errors.New("test error")
	formattedErr := formatErrorString(err, operations, "test", "true", "id", "1")

	expected := fmt.Sprintf("%v {operations: %v, test: true, id: 1}", err.Error(), operations)
	assert.Equal(expected, formattedErr)
}

func TestFormatErrorString_NilErrorWithData(t *testing.T) {
	assert := assert.New(t)

	operations := []Operation{"eventsource.error_test"}
	formattedErr := formatErrorString(nil, operations, "test", "true", "id", "1")

	expected := fmt.Sprintf("{operations: %v, test: true, id: 1}", operations)
	assert.Equal(expected, formattedErr)
}

func TestFormatErrorString_NilErrorWithoutData(t *testing.T) {
	assert := assert.New(t)

	operations := []Operation{"eventsource.error_test"}
	formattedErr := formatErrorString(nil, operations)

	assert.Equal("", formattedErr)
}

func TestFormatErrorString_WithMissingValue(t *testing.T) {
	assert := assert.New(t)

	operations := []Operation{"eventsource.error_test"}
	formattedErr := formatErrorString(nil, operations, "test", "true", "id")

	expected := fmt.Sprintf("{operations: %v, test: true, id: nil}", operations)
	assert.Equal(expected, formattedErr)
}
