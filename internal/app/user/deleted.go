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
	return &Deleted{
		ApplierModel: *eventsource.NewApplierModel(id, eventType, version, nil),
	}
}

type DeletedPayload struct {
	DeletedAt time.Time
}

// Apply implements the applier interface
func (event *Deleted) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	p, errDeserialize := event.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	u.DeletedAt = &p.DeletedAt
	return nil
}

func (event *Deleted) SetSerializedPayload(payload interface{}) error {
	deletedPayload, ok := payload.(DeletedPayload)
	if !ok {
		return eventsource.NewInvalidPayloadError(event.EventType, deletedPayload)
	}
	return event.Serialize(deletedPayload)
}

func (event *Deleted) GetDeserializedPayload() (*DeletedPayload, error) {
	var payload DeletedPayload
	errPayload := event.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if payload.DeletedAt.IsZero() {
		return nil, newPayloadMissingFieldsError(event.EventType, payload)
	}

	return &payload, nil
}
