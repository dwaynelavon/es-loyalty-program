package loyalty

import "github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"

// CreateUser command
type CreateUser struct {
	eventsource.CommandModel
	Username string `json:"username"`
}

// DeleteUser command
type DeleteUser struct {
	eventsource.CommandModel
	Username string `json:"username"`
}
