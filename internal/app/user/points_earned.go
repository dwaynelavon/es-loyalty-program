package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

// PointsEarned event is fired when a user action results in points earned
type PointsEarned struct {
	eventsource.ApplierModel
}

func NewPointsEarnedApplier(id, eventType string, version int) eventsource.Applier {
	return &PointsEarned{
		ApplierModel: *eventsource.NewApplierModel(id, eventType, version, nil),
	}
}

type PointsEarnedPayload struct {
	PointsEarned uint32 `json:"pointsEarned,omitempty"`
}

// Apply implements the applier interface
func (event *PointsEarned) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	p, errDeserialize := event.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}
	if eventsource.IsZero(p.PointsEarned) {
		return errors.New("EarnPoints event must have a non-zero point value")
	}

	u.Points += p.PointsEarned
	u.Version = event.Version
	return nil
}

func (event *PointsEarned) SetSerializedPayload(payload interface{}) error {
	pointsEarnedEvent, ok := payload.(PointsEarnedPayload)
	if !ok {
		return eventsource.NewInvalidPayloadError(event.EventType, pointsEarnedEvent)
	}
	return event.Serialize(pointsEarnedEvent)
}

func (event *PointsEarned) GetDeserializedPayload() (*PointsEarnedPayload, error) {
	var payload PointsEarnedPayload
	errPayload := event.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsZero(payload.PointsEarned) {
		return nil, newPayloadMissingFieldsError(event.EventType, payload)
	}

	return &payload, nil
}
