package user

import (
	"encoding/json"
	"time"
)

// Payload represents extra data that can be store with the User events
type Payload struct {
	Username  string     `json:"username"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
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
