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
func (applier *ReferralCreated) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	p, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	status, errStatus := GetReferralStatus(&p.ReferralStatus)
	if errStatus != nil {
		return errors.New("invalid status provided to ReferralCreated applier")
	}

	referral := Referral{
		ID:                p.ReferralID,
		ReferralCode:      p.ReferralCode,
		ReferredUserEmail: p.ReferredUserEmail,
		Status:            status,
	}

	u.Version = applier.Version
	u.Referrals = append(u.Referrals, referral)
	u.ReferralCode = &p.ReferralCode

	return nil
}

func (applier *ReferralCreated) SetSerializedPayload(payload interface{}) error {
	var operation eventsource.Operation = "user.ReferralCreated.SetSerializedPayload"

	referralCreatedPayload, ok := payload.(ReferralCreatedPayload)
	if !ok {
		return applier.PayloadErr(
			operation,
			payload,
		)
	}

	return applier.Serialize(referralCreatedPayload)
}

func (applier *ReferralCreated) GetDeserializedPayload() (*ReferralCreatedPayload, error) {
	var operation eventsource.Operation = "user.ReferralCreated.GetDeserializedPayload"

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
			operation,
			payload,
		)
	}

	return &payload, nil
}
