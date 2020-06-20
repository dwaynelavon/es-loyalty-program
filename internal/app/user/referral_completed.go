package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
)

// ReferralCompleted event is fired when a new user referral is created
type ReferralCompleted struct {
	eventsource.ApplierModel
}

type ReferralCompletedPayload struct {
	ReferralID string `json:"referralId,omitempty"`
}

func NewReferralCompletedApplier(id, eventType string, version int) eventsource.Applier {
	return &ReferralCompleted{
		ApplierModel: *eventsource.NewApplierModel(id, eventType, version, nil),
	}
}

// Apply implements the applier interface
func (applier *ReferralCompleted) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	p, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	for _, v := range u.Referrals {
		if v.ID == p.ReferralID {
			v.Status = ReferralStatusCreated
		}
	}

	u.Version = applier.Version
	return nil
}

func (applier *ReferralCompleted) SetSerializedPayload(payload interface{}) error {
	var operation eventsource.Operation = "user.ReferralCompleted.SetSerializedPayload"

	referralCompletedPayload, ok := payload.(ReferralCompletedPayload)
	if !ok {
		return applier.PayloadErr(
			operation,
			payload,
		)
	}
	return applier.Serialize(referralCompletedPayload)
}

func (applier *ReferralCompleted) GetDeserializedPayload() (*ReferralCompletedPayload, error) {
	var operation eventsource.Operation = "user.ReferralCompleted.GetDeserializedPayload"

	var payload ReferralCompletedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsAnyStringEmpty(&payload.ReferralID) {
		return nil, applier.PayloadErr(
			operation,
			payload,
		)
	}

	return &payload, nil
}
