package event

import (
	"context"
	"sort"

	"cloud.google.com/go/firestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/pkg/errors"
)

type store struct {
	firestoreClient *firestore.Client
}

// NewStore instantiates a new instance of the EventRepo
func NewStore(firestoreClient *firestore.Client) eventsource.EventStore {
	return &store{
		firestoreClient: firestoreClient,
	}
}

var eventCollection = "events"

func (s *store) Save(ctx context.Context, events ...eventsource.Event) error {
	if len(events) == 0 {
		return nil
	}

	sortedEvents := events
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].Version < sortedEvents[j].Version
	})

	batch := s.firestoreClient.Batch()
	ref := s.firestoreClient.Collection(eventCollection)
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
		return err
	}

	return nil
}

func (s *store) Load(
	ctx context.Context,
	aggregateID string,
	afterVersion int,
) (eventsource.History, error) {

	docs, errQuery := s.firestoreClient.
		Collection(eventCollection).
		OrderBy("version", firestore.Asc).
		Where("aggregateId", "==", aggregateID).
		Where("version", ">", afterVersion).
		Documents(ctx).
		GetAll()

	if errQuery != nil {
		return nil, errQuery
	}
	return transformDocumentsToHistory(docs)
}

func transformDocumentsToHistory(
	docs []*firestore.DocumentSnapshot,
) (eventsource.History, error) {
	history := make(eventsource.History, len(docs))
	for i, v := range docs {
		var record eventsource.Event
		err := v.DataTo(&record)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"unable to transform snapshot %v into History item",
				v.Ref.ID,
			)
		}
		history[i] = record
	}
	return history, nil
}
