package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/dwaynelavon/es-loyalty-program/graph/generated"
	"github.com/dwaynelavon/es-loyalty-program/graph/model"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/user"
)

func (r *mutationResolver) UserCreate(ctx context.Context, username string, email string, referredByCode *string) (*model.UserCreateResponse, error) {
	id := eventsource.NewUUID()

	err := r.Dispatcher.Dispatch(context.Background(), &loyalty.CreateUser{
		CommandModel: eventsource.CommandModel{
			ID: id,
		},
		Username:       username,
		Email:          email,
		ReferredByCode: referredByCode,
	})
	if err != nil {
		return nil, err
	}

	return &model.UserCreateResponse{
		Username: &username,
		UserID:   &id,
		Email:    &email,
	}, nil
}

func (r *mutationResolver) UserDelete(ctx context.Context, userID string) (*model.UserDeleteResponse, error) {
	err := r.Dispatcher.Dispatch(context.Background(), &loyalty.DeleteUser{
		CommandModel: eventsource.CommandModel{
			ID: userID,
		},
	})
	if err != nil {
		return nil, err
	}
	return &model.UserDeleteResponse{
		UserID: &userID,
	}, nil
}

func (r *mutationResolver) UserReferralCreate(ctx context.Context, userID string, referredUserEmail string) (*model.UserReferralCreatedResponse, error) {
	err := r.Dispatcher.Dispatch(ctx, &loyalty.CreateReferral{
		CommandModel: eventsource.CommandModel{
			ID: userID,
		},
		ReferredUserEmail: referredUserEmail,
	})
	if err != nil {
		return nil, err
	}
	return &model.UserReferralCreatedResponse{
		UserID:            &userID,
		ReferredUserEmail: &referredUserEmail,
	}, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]user.DTO, error) {
	return r.UserReadModel.Users(ctx)
}

func (r *userResolver) Points(ctx context.Context, obj *user.DTO) (int, error) {
	return int(obj.Points), nil
}

func (r *userResolver) Referrals(ctx context.Context, obj *user.DTO) ([]user.Referral, error) {
	return r.UserReadModel.Referrals(ctx, obj.UserID)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
