package user

import (
	"encoding/json"
	"time"
)

// Payload represents extra data that can be store with the User events
type Payload struct {
	Username  *string    `json:"username,omitempty"`
	Email     *string    `json:"email,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

func deserialize(payload *string) (*Payload, error) {
	if payload == nil {
		return nil, nil
	}
	var p Payload
	err := json.Unmarshal([]byte(*payload), &p)
	return &p, err
}

func serialize(payload *Payload) ([]byte, error) {
	if payload == nil {
		return nil, nil
	}
	return json.Marshal(payload)
}
