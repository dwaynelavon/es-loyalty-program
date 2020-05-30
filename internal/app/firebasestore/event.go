package firebasestore

import (
	"context"
	"sort"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

type store struct {
	firebaseApp *firebase.App
}

// NewStore instantiates a new instance of the EventRepo
func NewStore(firebaseApp *firebase.App) eventsource.EventStore {
	return &store{
		firebaseApp: firebaseApp,
	}
}

var eventCollection = "events"

func (s *store) Save(ctx context.Context, events ...eventsource.Event) error {
	if len(events) == 0 {
		return nil
	}

	client, errClient := s.firebaseApp.Firestore(ctx)
	if errClient != nil {
		return errors.Wrap(errClient, "error occured while creating Firestore client")
	}
	defer client.Close()

	sortedEvents := events
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].Version < sortedEvents[j].Version
	})

	batch := client.Batch()
	ref := client.Collection(eventCollection)
	for _, v := range events {
		m := &map[string]interface{}{
			"aggregateId": v.AggregateID,
			"version":     v.Version,
			"at":          v.EventAt,
			"payload":     v.Payload,
			"eventType":   v.EventType,
		}

		batch.Create(
			ref.NewDoc(),
			m,
		)
	}

	_, err := batch.Commit(ctx)
	if err != nil {
		return errors.Wrapf(
			err,
			"error occured in store while trying to commit a batch of %v events for aggregate: %v",
			len(events),
			events[0].AggregateID,
		)
	}

	return nil
}

func (s *store) Load(ctx context.Context, aggregateID string) (eventsource.History, error) {
	client, errClient := s.firebaseApp.Firestore(ctx)
	if errClient != nil {
		return nil, errors.Wrap(errClient, "error occured while creating Firestore client")
	}
	defer client.Close()

	docs, errQuery := client.
		Collection(eventCollection).
		OrderBy("version", firestore.Asc).
		Where("aggregateId", "==", aggregateID).
		Documents(ctx).
		GetAll()

	if errQuery != nil {
		return nil, errors.Wrapf(
			errQuery,
			"error occured while trying to load events for aggregateID: %v",
			aggregateID,
		)
	}
	return transformDocumentsToHistory(docs)
}

func transformDocumentsToHistory(docs []*firestore.DocumentSnapshot) (eventsource.History, error) {
	history := make(eventsource.History, len(docs))
	for i, v := range docs {
		var record eventsource.Event
		err := v.DataTo(&record)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error occured while transforming Firestore document: %v into event record",
				v.Ref.ID,
			)
		}
		history[i] = record
	}
	return history, nil
}
