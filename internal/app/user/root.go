package user

import (
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

var (
	userDeletedEventType           = "UserDeleted"
	userCreatedEventType           = "UserCreated"
	userReferralCreatedEventType   = "UserReferralCreated"
	userReferralCompletedEventType = "UserReferralCompleted"
	pointsEarnedEventType          = "PointsEarned"
)

// User encapsulates account information about an application user
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Version  int    `json:"version"`
	Username string `json:"username"`
	Points   uint32 `json:"points"`
	// TODO: should this be a pointer?
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
		a, err := getApplier(h)
		if err != nil {
			return err
		}

		errApply := a.Apply(u)
		if errApply != nil {
			return errApply
		}
	}
	return nil
}

/* ---------- helpers ---------- */
var (
	errInvalidAggregateType = errors.New("aggregate is not of type user.User")
)

func assertUserAggregate(u eventsource.Aggregate) (*User, error) {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return nil, errInvalidAggregateType
	}
	return user, nil
}

func newPayloadMissingFieldsError(eventType string, payload interface{}) error {
	return errors.Wrap(
		eventsource.NewInvalidPayloadError(eventType, payload),
		"missing required fields",
	)
}
