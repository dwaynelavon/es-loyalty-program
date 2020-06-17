package user

import (
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

func GetApplier(event eventsource.Event) (eventsource.Applier, error) {
	switch event.EventType {
	case PointsEarnedEventType:
		return &PointsEarned{
			ApplierModel: eventsource.ApplierModel{
				Event: event,
			},
		}, nil

	case UserCreatedEventType:
		return &Created{
			ApplierModel: eventsource.ApplierModel{
				Event: event,
			},
		}, nil

	case UserReferralCreatedEventType:
		return &ReferralCreated{
			ApplierModel: eventsource.ApplierModel{
				Event: event,
			},
		}, nil

	case UserDeletedEventType:
		return &Deleted{
			ApplierModel: eventsource.ApplierModel{
				Event: event,
			},
		}, nil

	case UserReferralCompletedEventType:
		return &ReferralCompleted{
			ApplierModel: eventsource.ApplierModel{
				Event: event,
			},
		}, nil

	default:
		return nil, errors.New("no registered applier for event type")
	}
}

func newPayloadMissingFieldsError(eventType string, payload interface{}) error {
	return errors.Wrap(
		eventsource.NewInvalidPayloadError(eventType, payload),
		"missing required fields",
	)
}
