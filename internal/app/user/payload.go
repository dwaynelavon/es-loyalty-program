package user

import (
	"encoding/json"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

// Payload represents extra data that can be store with the User events
type Payload struct {
	Username          *string    `json:"username,omitempty"`
	Email             *string    `json:"email,omitempty"`
	ReferralID        *string    `json:"referralId,omitempty"`
	ReferralCode      *string    `json:"referralCode,omitempty"`
	ReferredByCode    *string    `json:"referredByCode,omitempty"`
	ReferredUserEmail *string    `json:"referredUser,omitempty"`
	ReferralStatus    *string    `json:"referralStatus,omitempty"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
}

func deserialize(event eventsource.Event) (*Payload, error) {
	var (
		serializedPayload = event.Payload
		eventType         = event.EventType
	)
	if serializedPayload == nil {
		return nil, nil
	}
	var p Payload
	err := json.Unmarshal([]byte(*serializedPayload), &p)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error occurred while trying to deserialize payload for event type: %v",
			eventType,
		)
	}
	return &p, err
}

func serialize(commandType string, payload *Payload) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}
	serializedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error occurred while trying to serialize payload for command type: %v",
			commandType,
		)
	}
	return serializedPayload, nil
}
