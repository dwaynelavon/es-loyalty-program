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

// IsStringEmpty indicates whether or not a string is empty
func IsStringEmpty(str *string) bool {
	if str == nil || *str == "" {
		return true
	}
	return false
}

// IsAnyStringEmpty indicates whether or not any string in a slice is empty
func IsAnyStringEmpty(str ...*string) bool {
	for _, v := range str {
		if IsStringEmpty(v) {
			return true
		}
	}
	return false
}

// IsZero indicates whether or not a int is empty
func IsZero(i interface{}) bool {
	switch v := i.(type) {
	case *int:
		return v == nil || *v == 0
	case *int32:
		return v == nil || *v == 0
	case *int64:
		return v == nil || *v == 0
	case *uint32:
		return v == nil || *v == 0
	case *float32:
		return v == nil || *v == 0
	case *float64:
		return v == nil || *v == 0
	default:
		return false
	}
}

// StringToPointer takes a string and returns a pointer to the string value
func StringToPointer(str string) *string {
	return &str
}

// switch v := i.(type) {
// case *int, *int32, *int64, *uint32, *float32, *float64:
// 	return v == nil || *v == 0
// default:
// 	return false
// }
