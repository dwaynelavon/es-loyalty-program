package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

// PointsEarned event is fired when a user action results in points earned
type PointsEarned struct {
	eventsource.ApplierModel
}

func NewPointsEarnedApplier(
	id, eventType string,
	version int,
) eventsource.Applier {
	event := eventsource.NewEvent(id, eventType, version, nil)
	return &PointsEarned{ApplierModel: *eventsource.NewApplierModel(*event)}
}

type PointsEarnedPayload struct {
	PointsEarned uint32 `json:"pointsEarned,omitempty"`
}

// Apply implements the applier interface
func (applier *PointsEarned) Apply(agg eventsource.Aggregate) error {
	userAggregate, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	payload, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}
	if eventsource.IsZero(payload.PointsEarned) {
		return errors.New("EarnPoints event must have a non-zero point value")
	}

	userAggregate.Points += payload.PointsEarned
	userAggregate.Version = applier.Version
	return nil
}

func (applier *PointsEarned) SetSerializedPayload(payload interface{}) error {
	pointsEarnedEvent, ok := payload.(PointsEarnedPayload)
	if !ok {
		return applier.PayloadErr("user.PointsEarned.SetSerializedPayload", payload)
	}
	return applier.Serialize(pointsEarnedEvent)
}

func (applier *PointsEarned) GetDeserializedPayload() (
	*PointsEarnedPayload,
	error,
) {
	var payload PointsEarnedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if eventsource.IsZero(payload.PointsEarned) {
		return nil, applier.PayloadErr(
			"user.PointsEarned.GetDeserializedPayload",
			payload,
		)
	}

	return &payload, nil
}
