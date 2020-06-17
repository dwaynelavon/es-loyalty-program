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
func (event *ReferralCompleted) Apply(u eventsource.Aggregate) error {
	user, err := assertUserAggregate(u)
	if err != nil {
		return err
	}

	p, errDeserialize := event.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	for _, v := range user.Referrals {
		if v.ID == p.ReferralID {
			v.Status = ReferralStatusCompleted
		}
	}

	user.Version = event.Version
	return nil
}

func (event *ReferralCompleted) SetSerializedPayload(payload interface{}) error {
	referralCompletedPayload, ok := payload.(ReferralCompletedPayload)
	if !ok {
		return eventsource.NewInvalidPayloadError(event.EventType, referralCompletedPayload)
	}
	return event.Serialize(referralCompletedPayload)
}

func (event *ReferralCompleted) GetDeserializedPayload() (*ReferralCompletedPayload, error) {
	var payload ReferralCompletedPayload
	errPayload := event.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsAnyStringEmpty(&payload.ReferralID) {
		return nil, newPayloadMissingFieldsError(event.EventType, payload)
	}

	return &payload, nil
}
