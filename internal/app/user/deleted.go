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
func (applier *Deleted) Apply(agg eventsource.Aggregate) error {
	u, err := AssertUserAggregate(agg)
	if err != nil {
		return err
	}

	p, errDeserialize := applier.GetDeserializedPayload()
	if errDeserialize != nil {
		return errDeserialize
	}

	u.DeletedAt = &p.DeletedAt
	return nil
}

func (applier *Deleted) SetSerializedPayload(payload interface{}) error {
	var operation eventsource.Operation = "user.Deleted.SetSerializedPayload"

	deletedPayload, ok := payload.(DeletedPayload)
	if !ok {
		return applier.PayloadErr(
			operation,
			payload,
		)
	}
	return applier.Serialize(deletedPayload)
}

func (applier *Deleted) GetDeserializedPayload() (*DeletedPayload, error) {
	var operation eventsource.Operation = "user.Deleted.GetDeserializedPayload"

	var payload DeletedPayload
	errPayload := applier.Deserialize(&payload)
	if errPayload != nil {
		return nil, errPayload
	}

	if payload.DeletedAt.IsZero() {
		return nil, applier.PayloadErr(
			operation,
			payload,
		)
	}

	return &payload, nil
}
