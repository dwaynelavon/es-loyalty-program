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

func (s *userStore) CreateUser(ctx context.Context, user user.UserDTO) error {
	_, err := s.
		getUserDoc(user.UserID).
		Set(ctx, user)

	return err
}

func (s *userStore) CreateReferral(ctx context.Context, userID string, referral user.Referral) error {
	_, err := s.
		getUserReferralCollection(userID).
		Doc(referral.ID).
		Create(ctx, referral)

	return err
}

func (s *userStore) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.
		getUserDoc(userID).
		Delete(ctx)

	return err
}

func (s *userStore) Users(ctx context.Context) ([]user.UserDTO, error) {
	docs, err := s.
		getUserCollection().
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	return transformSnapshotsToUserDTOs(docs)
}

func (s *userStore) UpdateReferralStatus(ctx context.Context, userID string, referralID string, status user.ReferralStatus) error {
	_, err := s.
		getUserReferralCollection(userID).
		Doc(referralID).
		Update(ctx, []firestore.Update{
			{
				Path: "status", Value: string(status),
			},
		})

	return err
}

func (s *userStore) UserByReferralCode(ctx context.Context, referralCode string) (*user.UserDTO, error) {
	doc, err := s.
		getUserCollection().
		Where("referralCode", "==", referralCode).
		Documents(ctx).
		Next()

	if err != nil {
		return nil, err
	}

	return transformSnapshotToUserDTO(doc)
}

func (s *userStore) Referrals(ctx context.Context, userID string) ([]user.Referral, error) {
	docs, err := s.
		getUserReferralCollection(userID).
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	return transformSnapshotsToReferrals(docs)
}

/* ----- helpers ----- */
func (s *userStore) getUserCollection() *firestore.CollectionRef {
	return s.firestoreClient.
		Collection("users_es")
}

func (s *userStore) getUserReferralCollection(userID string) *firestore.CollectionRef {
	return s.
		getUserDoc(userID).
		Collection("referrals")
}

func (s *userStore) getUserDoc(userID string) *firestore.DocumentRef {
	return s.
		getUserCollection().
		Doc(userID)
}

func transformSnapshotToUserDTO(snapshot *firestore.DocumentSnapshot) (*user.UserDTO, error) {
	var user user.UserDTO
	err := snapshot.DataTo(&user)
	return &user, err
}

func transformSnapshotsToUserDTOs(snapshots []*firestore.DocumentSnapshot) ([]user.UserDTO, error) {
	users := []user.UserDTO{}
	var err error
	for _, v := range snapshots {
		var user user.UserDTO
		err = v.DataTo(&user)
		users = append(users, user)
	}
	return users, err
}

func transformSnapshotsToReferrals(snapshots []*firestore.DocumentSnapshot) ([]user.Referral, error) {
	referrals := []user.Referral{}
	var err error
	for _, v := range snapshots {
		var referral user.Referral
		err = v.DataTo(&referral)
		referrals = append(referrals, referral)
	}
	return referrals, err
}
