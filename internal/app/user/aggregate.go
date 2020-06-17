package user

import (
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

var (
	errInvalidAggregateType = errors.New("aggregate is not of type user.User")

	UserDeletedEventType           = "UserDeleted"
	UserCreatedEventType           = "UserCreated"
	UserReferralCreatedEventType   = "UserReferralCreated"
	UserReferralCompletedEventType = "UserReferralCompleted"
	PointsEarnedEventType          = "PointsEarned"
)

// ReferralStatus represents the state of a referral
type ReferralStatus string

const (
	ReferralStatusCreated   ReferralStatus = "Created"
	ReferralStatusSent      ReferralStatus = "Sent"
	ReferralStatusCompleted ReferralStatus = "Completed"
)

func GetReferralStatus(status *string) (ReferralStatus, error) {
	errInvalidStatus := errors.New("invalid referral status")
	if status == nil {
		return "", errInvalidStatus
	}
	switch *status {
	case string(ReferralStatusCreated):
		return ReferralStatusCreated, nil
	case string(ReferralStatusSent):
		return ReferralStatusSent, nil
	case string(ReferralStatusCompleted):
		return ReferralStatusCompleted, nil
	default:
		return "", errInvalidStatus
	}
}

//Referral is the struct that represents a new user's referral status
type Referral struct {
	ID                string         `json:"id" firestore:"id"`
	ReferralCode      string         `json:"referralCode" firestore:"referralCode"`
	ReferredUserEmail string         `json:"referredUserEmail" firestore:"referredUserEmail"`
	Status            ReferralStatus `json:"status" firestore:"status"`
	CreatedAt         time.Time      `json:"createdAt" firestore:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt" firestore:"updatedAt"`
}

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
		a, err := GetApplier(h)
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

func AssertUserAggregate(agg eventsource.Aggregate) (*User, error) {
	var u *User
	ok := false
	if u, ok = agg.(*User); !ok {
		return nil, errInvalidAggregateType
	}
	return u, nil
}
