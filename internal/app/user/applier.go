package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

var (
	userDeletedEventType           = "UserDeleted"
	userCreatedEventType           = "UserCreated"
	userReferralCreatedEventType   = "UserReferralCreated"
	userReferralCompletedEventType = "UserReferralCompletedEventType"

	errInvalidAggregateType = errors.New("aggregate is not of type user.User")
)

// TODO: Should these apply methods be used in the event handlers also

/* ---------- created ---------- */

// Created event is fired when a new user is created
type Created struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *Created) Apply(u eventsource.Aggregate) error {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return errInvalidAggregateType
	}

	p, errDeserialize := deserialize(event.Event)
	if errDeserialize != nil {
		return errDeserialize
	}

	if p.Email == nil || p.Username == nil || p.ReferralCode == nil {
		return newInvalidPayloadError(event.EventType)
	}
	user.Version = event.Version
	user.Username = *p.Username
	user.Email = *p.Email
	user.ReferralCode = p.ReferralCode
	return nil
}

/* ---------- referral created ---------- */

// ReferralCreated event is fired when a new user referral is created
type ReferralCreated struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *ReferralCreated) Apply(u eventsource.Aggregate) error {
	user, err := assertUserAggregate(u)
	if err != nil {
		return err
	}
	p, errDeserialize := deserialize(event.Event)
	if errDeserialize != nil {
		return errDeserialize
	}
	status, errStatus := getReferralStatus(p.ReferralStatus)
	if errStatus != nil ||
		eventsource.IsStringEmpty(p.ReferredUserEmail) ||
		eventsource.IsStringEmpty(p.ReferralCode) {
		return newInvalidPayloadError(event.EventType)
	}

	referral := Referral{
		ID:                *p.ReferralID,
		ReferralCode:      *p.ReferralCode,
		ReferredUserEmail: *p.ReferredUserEmail,
		Status:            status,
	}

	user.Version = event.Version
	user.Referrals = append(user.Referrals, referral)
	user.ReferralCode = p.ReferralCode

	return nil
}

/* ---------- referral completed ---------- */

// ReferralCompleted event is fired when a new user referral is created
type ReferralCompleted struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *ReferralCompleted) Apply(u eventsource.Aggregate) error {
	user, err := assertUserAggregate(u)
	if err != nil {
		return err
	}
	p, errDeserialize := deserialize(event.Event)
	if errDeserialize != nil {
		return errDeserialize
	}
	if eventsource.IsStringEmpty(p.ReferralID) {
		return errors.New("ReferralCompleted event must have a ReferralID")
	}

	for _, v := range user.Referrals {
		if v.ID == *p.ReferralID {
			v.Status = ReferralStatusCompleted
		}
	}

	user.Version = event.Version
	return nil
}

/* ---------- deleted ---------- */

// Deleted event is fired when a user is deleted
type Deleted struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *Deleted) Apply(u eventsource.Aggregate) error {
	user, err := assertUserAggregate(u)
	if err != nil {
		return err
	}
	p, errDeserialize := deserialize(event.Event)
	if errDeserialize != nil {
		return errDeserialize
	}
	user.DeletedAt = p.DeletedAt
	return nil
}
