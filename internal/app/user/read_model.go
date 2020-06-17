package user

import (
	"context"
	"time"

	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"

	"go.uber.org/zap"
)

type DTO struct {
	eventsource.AggregateBase
	UserID         string    `json:"userId" firestore:"userId"`
	Username       string    `json:"username" firestore:"username"`
	Email          string    `json:"email" firestore:"email"`
	Points         uint32    `json:"points" firestore:"points"`
	ReferredByCode *string   `json:"referredByCode" firestore:"referredByCode"`
	ReferralCode   string    `json:"referralCode" firestore:"referralCode"`
	CreatedAt      time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// EventVersion returns the last event version processed
func (u *DTO) EventVersion() int {
	return u.Version
}

type ReadModel interface {
	Users(context.Context) ([]DTO, error)
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

func (r *readModel) Users(ctx context.Context) ([]DTO, error) {
	return r.readRepo.Users(ctx)
}

func (r *readModel) Referrals(ctx context.Context, userID string) ([]Referral, error) {
	return r.readRepo.Referrals(ctx, userID)
}
