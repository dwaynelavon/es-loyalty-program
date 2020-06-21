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

func (s *userStore) CreateUser(
	ctx context.Context,
	user user.DTO,
) error {
	_, err := s.
		getUserDoc(user.UserID).
		Set(ctx, user)

	return err
}

func (s *userStore) CreateReferral(
	ctx context.Context,
	userID string,
	referral user.Referral,
	version int,
) error {
	referralRef := s.getUserReferralCollection(userID).Doc(referral.ID)
	batch := s.
		batchUpdateWithVersion(userID, version).
		Create(referralRef, referral)

	_, err := batch.Commit(ctx)
	return err
}

func (s *userStore) DeleteUser(
	ctx context.Context,
	userID string,
) error {
	// TODO: Is soft delete more appropriate here
	// If using soft delete, need to update version also
	_, err := s.
		getUserDoc(userID).
		Delete(ctx)

	return err
}

func (s *userStore) Users(ctx context.Context) ([]user.DTO, error) {
	docs, err := s.
		getUserCollection().
		Documents(ctx).
		GetAll()

	if err != nil {
		return nil, err
	}

	return transformSnapshotsToUserDTOs(docs)
}

func (s *userStore) UpdateReferralStatus(
	ctx context.Context,
	userID, referralID string,
	status user.ReferralStatus,
	version int,
) error {
	referralRef := s.getUserReferralCollection(userID).Doc(referralID)
	batch := s.
		batchUpdateWithVersion(userID, version).
		Update(referralRef, []firestore.Update{
			{Path: "status", Value: string(status)},
		})

	_, err := batch.Commit(ctx)
	return err
}

func (s *userStore) EarnPoints(
	ctx context.Context,
	userID string,
	points uint32,
	version int,
) error {
	_, err := s.
		getUserDoc(userID).
		Update(ctx, []firestore.Update{
			{Path: "points", Value: firestore.Increment(points)},
			{Path: "version", Value: version},
		})

	return err
}

func (s *userStore) UserByReferralCode(
	ctx context.Context,
	referralCode string,
) (*user.DTO, error) {
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

func (s *userStore) User(
	ctx context.Context,
	userID string,
) (*user.DTO, error) {
	doc, err := s.
		getUserDoc(userID).
		Get(ctx)

	if err != nil {
		return nil, err
	}

	return transformSnapshotToUserDTO(doc)
}

func (s *userStore) Referrals(
	ctx context.Context,
	userID string,
) ([]user.Referral, error) {
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

func (s *userStore) batchUpdateWithVersion(
	userID string,
	version int,
) *firestore.WriteBatch {
	batch := s.firestoreClient.Batch()
	userRef := s.getUserDoc(userID)
	return batch.Update(
		userRef,
		[]firestore.Update{{Path: "version", Value: version}},
	)
}

func transformSnapshotToUserDTO(
	snapshot *firestore.DocumentSnapshot,
) (*user.DTO, error) {
	var user user.DTO
	err := snapshot.DataTo(&user)
	return &user, err
}

func transformSnapshotsToUserDTOs(
	snapshots []*firestore.DocumentSnapshot,
) ([]user.DTO, error) {
	users := []user.DTO{}
	var err error
	for _, v := range snapshots {
		var user user.DTO
		err = v.DataTo(&user)
		users = append(users, user)
	}
	return users, err
}

func transformSnapshotsToReferrals(
	snapshots []*firestore.DocumentSnapshot,
) ([]user.Referral, error) {
	referrals := []user.Referral{}
	var err error
	for _, v := range snapshots {
		var referral user.Referral
		err = v.DataTo(&referral)
		referrals = append(referrals, referral)
	}
	return referrals, err
}
