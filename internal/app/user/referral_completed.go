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

func NewReferralCompletedApplier(
	id, eventType string,
	version int,
) eventsource.Applier {
	event := eventsource.NewEvent(id, eventType, version, nil)
	return &ReferralCompleted{
		ApplierModel: *eventsource.NewApplierModel(*event),
	}
}

// Apply implements the applier interface
func (applier *ReferralCompleted) Apply(agg eventsource.Aggregate) error {
	userAggregate, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	payload, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	for _, v := range userAggregate.Referrals {
		if v.ID == payload.ReferralID {
			v.Status = ReferralStatusCreated
		}
	}

	userAggregate.Version = applier.Version
	return nil
}

func (applier *ReferralCompleted) SetSerializedPayload(
	payload interface{},
) error {
	referralCompletedPayload, ok := payload.(ReferralCompletedPayload)
	if !ok {
		return applier.PayloadErr(
			"user.ReferralCompleted.SetSerializedPayload",
			payload,
		)
	}
	return applier.Serialize(referralCompletedPayload)
}

func (applier *ReferralCompleted) GetDeserializedPayload() (
	*ReferralCompletedPayload,
	error,
) {
	var payload ReferralCompletedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsAnyStringEmpty(&payload.ReferralID) {
		return nil, applier.PayloadErr(
			"user.ReferralCompleted.GetDeserializedPayload",
			payload,
		)
	}

	return &payload, nil
}
