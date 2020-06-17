package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
)

// Created event is fired when a new user is created
type Created struct {
	eventsource.ApplierModel
}

func NewCreatedApplier(id, eventType string, version int) eventsource.Applier {
	return &Created{
		ApplierModel: *eventsource.NewApplierModel(id, eventType, version, nil),
	}
}

type CreatedPayload struct {
	Username       string  `json:"username,omitempty"`
	Email          string  `json:"email,omitempty"`
	ReferredByCode *string `json:"referredByCode,omitempty"`
	ReferralCode   string  `json:"referralCode,omitempty"`
}

// Apply implements the applier interface
func (event *Created) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	p, errDeserialize := event.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	u.Version = event.Version
	u.Username = p.Username
	u.Email = p.Email
	u.ReferralCode = &p.ReferralCode
	return nil
}

func (event *Created) SetSerializedPayload(payload interface{}) error {
	createdPayload, ok := payload.(CreatedPayload)
	if !ok {
		return eventsource.NewInvalidPayloadError(event.EventType, payload)
	}

	return event.Serialize(createdPayload)
}

func (event *Created) GetDeserializedPayload() (*CreatedPayload, error) {
	var payload CreatedPayload
	errPayload := event.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsAnyStringEmpty(
		&payload.Username,
		&payload.Email,
		&payload.ReferralCode,
	) {
		return nil, newPayloadMissingFieldsError(event.EventType, payload)
	}

	return &payload, nil
}
