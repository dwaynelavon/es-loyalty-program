package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
)

// Created event is fired when a new user is created
type Created struct {
	eventsource.ApplierModel
}

func NewCreatedApplier(id, eventType string, version int) eventsource.Applier {
	event := eventsource.NewEvent(id, eventType, version, nil)
	return &Created{
		ApplierModel: *eventsource.NewApplierModel(*event),
	}
}

type CreatedPayload struct {
	Username       string  `json:"username,omitempty"`
	Email          string  `json:"email,omitempty"`
	ReferredByCode *string `json:"referredByCode,omitempty"`
	ReferralCode   string  `json:"referralCode,omitempty"`
}

// Apply implements the applier interface
func (applier *Created) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	payload, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	u.Version = applier.Version
	u.Username = payload.Username
	u.Email = payload.Email
	u.ReferralCode = &payload.ReferralCode
	return nil
}

func (applier *Created) SetSerializedPayload(payload interface{}) error {
	var operation eventsource.Operation = "user.Created.SetSerializedPayload"

	createdPayload, ok := payload.(CreatedPayload)
	if !ok {
		return applier.PayloadErr(
			operation,
			payload,
		)
	}

	return applier.Serialize(createdPayload)
}

func (applier *Created) GetDeserializedPayload() (*CreatedPayload, error) {
	var payload CreatedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsAnyStringEmpty(
		&payload.Username,
		&payload.Email,
		&payload.ReferralCode,
	) {
		return nil, applier.PayloadErr(
			"user.Created.GetDeserializedPayload",
			payload,
		)
	}

	return &payload, nil
}
