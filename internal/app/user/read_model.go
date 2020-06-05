package user

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type UserDTO struct {
	UserId    string    `json:"userId" firestore:"userId"`
	Username  string    `json:"username" firestore:"username"`
	Email     string    `json:"email" firestore:"email"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type ReadRepo interface {
	CreateUser(context.Context, UserDTO) error
	DeleteUser(context.Context, string) error
	Users(context.Context) ([]UserDTO, error)
}

type ReadModel interface {
	Users(context.Context) ([]UserDTO, error)
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
