package user

import (
	"fmt"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

// User encapsulates account information about an application user
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Version  int    `json:"version"`
	Username string `json:"username"`
	// TODO: should this be a pointer
	ReferralCode *string `json:"referralCode"`
	// TODO: Should we include this here? Seems like it should go only in the DTO
	Referrals []Referral `json:"referrals"`
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
			return fmt.Errorf("error occurred while trying to get applier for: %v", h.EventType)
		}
		errApply := a.Apply(u)
		if errApply != nil {
			return errors.Wrapf(errApply, "error occurred while trying to apply %v", h.EventType)
		}
	}
	return nil
}

/* ---------- helpers ---------- */

func assertUserAggregate(u eventsource.Aggregate) (*User, error) {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return nil, errInvalidAggregateType
	}
	return user, nil
}

func newInvalidPayloadError(eventType string) error {
	return errors.Errorf("invalid payload provided to %v", eventType)
}
