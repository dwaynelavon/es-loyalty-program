package user

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ReferralStatus represents the state of a referral
type ReferralStatus string

const ReferralStatusCreated ReferralStatus = "Created"
const ReferralStatusSent ReferralStatus = "Sent"
const ReferralStatusCompleted ReferralStatus = "Completed"

func getReferralStatus(status *string) (ReferralStatus, error) {
	errInvalidStatus := errors.New("invalid referral status")
	if status == nil {
		return "", errInvalidStatus
	}
	switch *status {
	case string(ReferralStatusCreated):
		return ReferralStatusCreated, nil
	case string(ReferralStatusSent):
		return ReferralStatusSent, nil
	case string(ReferralStatusCompleted):
		return ReferralStatusCompleted, nil
	default:
		return "", errInvalidStatus
	}
}

//Referral is the struct that represents a new user's referral status
type Referral struct {
	ID                string         `json:"id" firestore:"id"`
	ReferralCode      string         `json:"referralCode" firestore:"referralCode"`
	ReferredUserEmail string         `json:"referredUserEmail" firestore:"referredUserEmail"`
	Status            ReferralStatus `json:"status" firestore:"status"`
	CreatedAt         time.Time      `json:"createdAt" firestore:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt" firestore:"updatedAt"`
}

type UserDTO struct {
	UserID       string    `json:"userId" firestore:"userId"`
	Username     string    `json:"username" firestore:"username"`
	Email        string    `json:"email" firestore:"email"`
	ReferralCode string    `json:"referralCode" firestore:"referralCode"`
	CreatedAt    time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type ReadRepo interface {
	CreateUser(context.Context, UserDTO) error
	CreateReferral(ctx context.Context, userID string, referral Referral) error
	UpdateReferralStatus(ctx context.Context, userID string, referralID string, status ReferralStatus) error
	DeleteUser(context.Context, string) error
	Users(context.Context) ([]UserDTO, error)
	Referrals(ctx context.Context, userID string) ([]Referral, error)
	UserByReferralCode(ctx context.Context, referralCode string) (*UserDTO, error)
}

type ReadModel interface {
	Users(context.Context) ([]UserDTO, error)
	Referrals(ctx context.Context, userID string) ([]Referral, error)
}

type readModel struct {
	readRepo ReadRepo
	logger   *zap.Logger
}

type ReadModelParams struct {
	ReadRepo ReadRepo
	Logger   *zap.Logger
}

func NewReadModel(params ReadModelParams) ReadModel {
	return &readModel{
		readRepo: params.ReadRepo,
		logger:   params.Logger,
	}
}

func (r *readModel) Users(ctx context.Context) ([]UserDTO, error) {
	return r.readRepo.Users(ctx)
}

func (r *readModel) Referrals(ctx context.Context, userID string) ([]Referral, error) {
	return r.readRepo.Referrals(ctx, userID)
}
