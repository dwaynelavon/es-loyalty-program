package readmodel

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
)

type userStore struct {
	firestoreClient *firestore.Client
}

// NewUserStore instantiates a new instance of the EventRepo
func NewUserStore(firestoreClient *firestore.Client) user.ReadRepo {
	return &userStore{
		firestoreClient: firestoreClient,
	}
}

var userCollection = "users_es"

func (s *userStore) CreateUser(ctx context.Context, user user.UserDTO) error {
	_, err := s.firestoreClient.
		Collection(userCollection).
		NewDoc().
		Set(ctx, user)

	return err
}

func (s *userStore) DeleteUser(ctx context.Context, userID string) error {
	doc, docErr := s.firestoreClient.
		Collection(userCollection).
		Where("userId", "==", userID).
		Documents(ctx).
		Next()
	if docErr != nil {
		return docErr
	}

	_, err := doc.Ref.Delete(ctx)

	return err
}

func (s *userStore) Users(ctx context.Context) ([]user.UserDTO, error) {
	docs, err := s.firestoreClient.
		Collection(userCollection).
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	return transformSnapshotsToUserDTOs(docs)
}

func transformSnapshotsToUserDTOs(docs []*firestore.DocumentSnapshot) ([]user.UserDTO, error) {
	users := []user.UserDTO{}
	var err error
	for _, v := range docs {
		var user user.UserDTO
		err = v.DataTo(&user)
		users = append(users, user)
	}
	return users, err
}
