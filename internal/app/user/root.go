package user

import "github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"

type EventStore interface {
	eventsource.EventStore
}
