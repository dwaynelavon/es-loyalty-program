package user

import (
	"fmt"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
)

// User encapsulates account information about an application user
type User struct {
	ID        string     `json:"id"`
	Version   int        `json:"version"`
	Username  string     `json:"username"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

// NewUser creates a new instance of the User aggregate
func NewUser(id string) eventsource.Aggregate {
	return &User{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// EventVersion returns the current event version
func (u *User) EventVersion() int {
	return u.Version
}

// Apply takes event history and applies them to an aggregate
func (u *User) Apply(history eventsource.History) error {
	for _, h := range history {
		a, ok := getApplier(h)
		if !ok {
			return fmt.Errorf("error occured while trying to get applier for: %v", h.EventType)
		}
		errApply := a.Apply(u)
		if errApply != nil {
			return fmt.Errorf("error occured while trying to apply: %v", h.EventType)
		}
	}
	return nil
}
