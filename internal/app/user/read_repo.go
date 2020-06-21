package user

import (
	"context"
)

type ReadRepo interface {
	CreateUser(ctx context.Context, user DTO) error
	CreateReferral(ctx context.Context, userID string, referral Referral, version int) error
	EarnPoints(ctx context.Context, userID string, points uint32, version int) error
	UpdateReferralStatus(ctx context.Context, userID string, referralID string, status ReferralStatus, version int) error
	DeleteUser(ctx context.Context, userID string) error
	Users(context.Context) ([]DTO, error)
	User(ctx context.Context, userID string) (*DTO, error)
	Referrals(ctx context.Context, userID string) ([]Referral, error)
	UserByReferralCode(ctx context.Context, referralCode string) (*DTO, error)
}
