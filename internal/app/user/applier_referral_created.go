package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

// ReferralCreated event is fired when a new user referral is created
type ReferralCreated struct {
	eventsource.ApplierModel
}

func NewReferralCreatedApplier(id, eventType string, version int) eventsource.Applier {
	return &ReferralCreated{
		ApplierModel: *eventsource.NewApplierModel(id, eventType, version, nil),
	}
}

type ReferralCreatedPayload struct {
	ReferredUserEmail string `json:"referredUserEmail,omitempty"`
	ReferralCode      string `json:"referralCode,omitempty"`
	ReferralID        string `json:"referralId,omitempty"`
	ReferralStatus    string `json:"referralStatus,omitempty"`
}

// Apply implements the applier interface
func (event *ReferralCreated) Apply(u eventsource.Aggregate) error {
	user, err := assertUserAggregate(u)
	if err != nil {
		return err
	}

	p, errDeserialize := event.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	status, errStatus := getReferralStatus(&p.ReferralStatus)
	if errStatus != nil {
		return errors.New("invalid status provided to ReferralCreated applier")
	}

	referral := Referral{
		ID:                p.ReferralID,
		ReferralCode:      p.ReferralCode,
		ReferredUserEmail: p.ReferredUserEmail,
		Status:            status,
	}

	user.Version = event.Version
	user.Referrals = append(user.Referrals, referral)
	user.ReferralCode = &p.ReferralCode

	return nil
}

func (event *ReferralCreated) SetSerializedPayload(payload interface{}) error {
	referralCreatedPayload, ok := payload.(ReferralCreatedPayload)
	if !ok {
		return eventsource.NewInvalidPayloadError(event.EventType, referralCreatedPayload)
	}

	return event.Serialize(referralCreatedPayload)
}

func (event *ReferralCreated) GetDeserializedPayload() (*ReferralCreatedPayload, error) {
	var payload ReferralCreatedPayload
	errPayload := event.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsAnyStringEmpty(
		&payload.ReferralID,
		&payload.ReferredUserEmail,
		&payload.ReferralCode,
		&payload.ReferralStatus,
	) {
		return nil, newPayloadMissingFieldsError(event.EventType, payload)
	}

	return &payload, nil
}
