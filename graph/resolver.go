package graph

import "github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Dispatcher eventsource.CommandDispatcher
}
