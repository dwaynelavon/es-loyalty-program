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
func (applier *PointsEarned) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	p, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}
	if eventsource.IsZero(p.PointsEarned) {
		return errors.New("EarnPoints event must have a non-zero point value")
	}

	u.Points += p.PointsEarned
	u.Version = applier.Version
	return nil
}

func (applier *PointsEarned) SetSerializedPayload(payload interface{}) error {
	var operation eventsource.Operation = "user.PointsEarned.SetSerializedPayload"

	pointsEarnedEvent, ok := payload.(PointsEarnedPayload)
	if !ok {
		return applier.PayloadErr(operation, payload)
	}
	return applier.Serialize(pointsEarnedEvent)
}

func (applier *PointsEarned) GetDeserializedPayload() (*PointsEarnedPayload, error) {
	var operation eventsource.Operation = "user.PointsEarned.GetDeserializedPayload"

	var payload PointsEarnedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsZero(payload.PointsEarned) {
		return nil, applier.PayloadErr(
			operation,
			payload,
		)
	}

	return &payload, nil
}
