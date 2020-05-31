package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/dwaynelavon/es-loyalty-program/graph/generated"
	"github.com/dwaynelavon/es-loyalty-program/graph/model"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/eventsource"
	"github.com/dwaynelavon/es-loyalty-program/internal/app/loyalty"
)

func (r *mutationResolver) UserCreate(ctx context.Context, username string) (*model.UserCreateResponse, error) {
	id := eventsource.NewUUID()
	err := r.Dispatcher.Dispatch(context.Background(), &loyalty.CreateUser{
		CommandModel: eventsource.CommandModel{
			ID: id,
		},
		Username: "admin",
	})
	if err != nil {
		return &model.UserCreateResponse{
			Status: model.StatusRejected,
		}, err
	}
	return &model.UserCreateResponse{
		Status:   model.StatusAccepted,
		Username: &username,
		UserID:   &id,
	}, nil
}

func (r *mutationResolver) UserDelete(ctx context.Context, userID string) (*model.UserDeleteResponse, error) {
	err := r.Dispatcher.Dispatch(context.Background(), &loyalty.DeleteUser{
		CommandModel: eventsource.CommandModel{
			ID: userID,
		},
	})
	if err != nil {
		return &model.UserDeleteResponse{
			Status: model.StatusRejected,
		}, err
	}
	return &model.UserDeleteResponse{
		Status: model.StatusAccepted,
		UserID: &userID,
	}, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }