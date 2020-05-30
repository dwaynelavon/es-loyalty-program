package eventsource

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

// typeOf is a convenience function that returns the name of a type
//
// This is used so commonly throughout the code that it is better to
// have this convenience function and also allows for changing the scheme
// used for the type name more easily if desired.
func typeOf(i interface{}) string {
	return reflect.TypeOf(i).Elem().Name()
}

// NewUUID returns a new v4 uuid as a string
func NewUUID() string {
	return uuid.New().String()
}

// Int returns a pointer to int.
//
// There are a number of places where a pointer to int
// is required such as expectedVersion argument on the repository
// and this helper function makes keeps the code cleaner in these
// cases.
func Int(i int) *int {
	return &i
}

func TimeNow() *time.Time {
	t := time.Now()
	return &t
}
