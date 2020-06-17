package user

import (
	"time"
)

// Payload represents extra data that can be store with the User events
type Payload struct {
	Username          *string    `json:"username,omitempty"`
	Email             *string    `json:"email,omitempty"`
	PointsEarned      *uint32    `json:"pointsEarned,omitempty"`
	ReferralID        *string    `json:"referralId,omitempty"`
	ReferralCode      *string    `json:"referralCode,omitempty"`
	ReferredByCode    *string    `json:"referredByCode,omitempty"`
	ReferredUserEmail *string    `json:"referredUser,omitempty"`
	ReferralStatus    *string    `json:"referralStatus,omitempty"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
}
