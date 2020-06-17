package user

import (
	"context"
)

type ReadRepo interface {
	CreateUser(context.Context, DTO) error
	CreateReferral(ctx context.Context, userID string, referral Referral) error
	EarnPoints(ctx context.Context, userID string, points uint32) error
	UpdateReferralStatus(ctx context.Context, userID string, referralID string, status ReferralStatus) error
	DeleteUser(context.Context, string) error
	Users(context.Context) ([]DTO, error)
	Referrals(ctx context.Context, userID string) ([]Referral, error)
	UserByReferralCode(ctx context.Context, referralCode string) (*DTO, error)
}
