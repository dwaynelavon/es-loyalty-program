package user

import (
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
)

// Deleted event is fired when a user is deleted
type Deleted struct {
	eventsource.ApplierModel
}

func NewDeletedApplier(id, eventType string, version int) eventsource.Applier {
	event := eventsource.NewEvent(id, eventType, version, nil)
	return &Deleted{ApplierModel: *eventsource.NewApplierModel(*event)}
}

type DeletedPayload struct {
	DeletedAt time.Time
}

// Apply implements the applier interface
func (applier *Deleted) Apply(agg eventsource.Aggregate) error {
	userAggregate, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	payload, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	userAggregate.DeletedAt = &payload.DeletedAt
	return nil
}

func (applier *Deleted) SetSerializedPayload(payload interface{}) error {
	deletedPayload, ok := payload.(DeletedPayload)
	if !ok {
		return applier.PayloadErr(
			"user.Deleted.SetSerializedPayload",
			payload,
		)
	}

	return applier.Serialize(deletedPayload)
}

func (applier *Deleted) GetDeserializedPayload() (*DeletedPayload, error) {
	var payload DeletedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if payload.DeletedAt.IsZero() {
		return nil, applier.PayloadErr(
			"user.Deleted.GetDeserializedPayload",
			payload,
		)
	}

	return &payload, nil
}
