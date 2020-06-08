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

	errInvalidAggregateType = errors.New("invalid aggregate type being applied to a User")
)

// Created event is fired when a new user is created
type Created struct {
	eventsource.Event
}

// TODO: Should these apply methods be used in the event handlers also
// Apply implements the applier interface
func (event *Created) Apply(u eventsource.Aggregate) error {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return errInvalidAggregateType
	}

	p, errDeserialize := deserialize(event.EventType, event.Payload)
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

// ReferralCreated event is fired when a new user referral is created
type ReferralCreated struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *ReferralCreated) Apply(u eventsource.Aggregate) error {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return errInvalidAggregateType
	}
	p, errDeserialize := deserialize(event.EventType, event.Payload)
	if errDeserialize != nil {
		return errDeserialize
	}

	status, errStatus := getReferralStatus(p.ReferralStatus)
	if errStatus != nil ||
		p.ReferredUserEmail == nil ||
		p.ReferralCode == nil {
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

// ReferralCompleted event is fired when a new user referral is created
type ReferralCompleted struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *ReferralCompleted) Apply(u eventsource.Aggregate) error {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return errInvalidAggregateType
	}
	p, errDeserialize := deserialize(event.EventType, event.Payload)
	if errDeserialize != nil {
		return errDeserialize
	}
	if p.ReferralID == nil {
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

// Deleted event is fired when a user is deleted
type Deleted struct {
	eventsource.Event
}

// Apply implements the applier interface
func (event *Deleted) Apply(u eventsource.Aggregate) error {
	var user *User
	ok := false
	if user, ok = u.(*User); !ok {
		return errInvalidAggregateType
	}
	p, errDeserialize := deserialize(event.EventType, event.Payload)
	if errDeserialize != nil {
		return errDeserialize
	}
	user.DeletedAt = p.DeletedAt
	return nil
}
