package loyalty

import "github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"

// CreateUser command
type CreateUser struct {
	eventsource.CommandModel
	Username       string  `json:"username"`
	Email          string  `json:"email"`
	ReferredByCode *string `json:"referredByCode"`
}

// DeleteUser command
type DeleteUser struct {
	eventsource.CommandModel
}

// CreateReferral command
type CreateReferral struct {
	eventsource.CommandModel
	ReferredUserEmail string `json:"referredUserEmail"`
}

// CompleteReferral command
type CompleteReferral struct {
	eventsource.CommandModel
	ReferredUserEmail string `json:"referredUserEmail"`
	ReferredUserID    string `json:"referredUserId"`
	ReferredByCode    string `json:"referredByCode"`
}

// EarnPoints command
type EarnPoints struct {
	eventsource.CommandModel
	Points uint32 `json:"points"`
}
