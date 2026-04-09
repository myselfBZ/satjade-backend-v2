package users_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/store"
)

type service struct{
	userStore store.UserStore
}

func New(userStore store.UserStore) UserService {
	return &service{userStore}
}

type UserService interface{
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) 
	Delete(ctx context.Context, id uuid.UUID) error
	GetMany(ctx context.Context) ([]domain.User, error) 
}

func (s *service) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userStore.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) (error) {
	return s.userStore.Delete(ctx, id)
}

func (s *service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userStore.GetByEmail(ctx, email)
}

func (s *service) GetMany(ctx context.Context) ([]domain.User, error) {
	return s.userStore.GetMany(ctx)
} 
