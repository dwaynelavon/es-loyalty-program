package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

// ReferralCreated event is fired when a new user referral is created
type ReferralCreated struct {
	eventsource.ApplierModel
}

func NewReferralCreatedApplier(
	id, eventType string,
	version int,
) eventsource.Applier {
	event := eventsource.NewEvent(id, eventType, version, nil)
	return &ReferralCreated{
		ApplierModel: *eventsource.NewApplierModel(*event),
	}
}

type ReferralCreatedPayload struct {
	ReferredUserEmail string `json:"referredUserEmail,omitempty"`
	ReferralCode      string `json:"referralCode,omitempty"`
	ReferralID        string `json:"referralId,omitempty"`
	ReferralStatus    string `json:"referralStatus,omitempty"`
}

// Apply implements the applier interface
func (applier *ReferralCreated) Apply(agg eventsource.Aggregate) error {
	userAggregate, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	payload, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	status, errStatus := GetReferralStatus(&payload.ReferralStatus)
	if errStatus != nil {
		return errors.New("invalid status provided to ReferralCreated applier")
	}

	referral := Referral{
		ID:                payload.ReferralID,
		ReferralCode:      payload.ReferralCode,
		ReferredUserEmail: payload.ReferredUserEmail,
		Status:            status,
	}

	userAggregate.Version = applier.Version
	userAggregate.Referrals = append(userAggregate.Referrals, referral)
	userAggregate.ReferralCode = &payload.ReferralCode

	return nil
}

func (applier *ReferralCreated) SetSerializedPayload(
	payload interface{},
) error {
	referralCreatedPayload, ok := payload.(ReferralCreatedPayload)
	if !ok {
		return applier.PayloadErr(
			"user.ReferralCreated.SetSerializedPayload",
			payload,
		)
	}

	return applier.Serialize(referralCreatedPayload)
}

func (applier *ReferralCreated) GetDeserializedPayload() (
	*ReferralCreatedPayload,
	error,
) {
	var payload ReferralCreatedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsAnyStringEmpty(
		&payload.ReferralID,
		&payload.ReferredUserEmail,
		&payload.ReferralCode,
		&payload.ReferralStatus,
	) {
		return nil, applier.PayloadErr(
			"user.ReferralCreated.GetDeserializedPayload",
			payload,
		)
	}

	return &payload, nil
}
