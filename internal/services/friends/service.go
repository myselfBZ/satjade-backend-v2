package friends_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/store"
)

type service struct {
	friendsStore store.FriendsStore
}

func New(friendsStore store.FriendsStore) FriendsService {
	return &service{friendsStore}
}

type CreateFriendshipRequest struct {
	*store.CreateFriendRequestParams
}

type FriendsService interface {
	GetManyByUser(ctx context.Context, userId uuid.UUID) ([]domain.Friend, error)
	Create(ctx context.Context, params *store.CreateFriendParams) (domain.Friend, error)
	Delete(ctx context.Context, friendshipId uuid.UUID) error
	AcceptFriendshipRequest(ctx context.Context, requestId uuid.UUID) (domain.Friendship, error)
	DeleteFriendshipRequest(ctx context.Context, id uuid.UUID) error
	GetFriendshipRequests(ctx context.Context, userId uuid.UUID) ([]domain.FriendshipRequest, error)
	CreateFriendshipRequest(ctx context.Context, params *CreateFriendshipRequest) (domain.FriendshipRequest, error)
}

func (s *service) CreateFriendshipRequest(ctx context.Context, params *CreateFriendshipRequest) (domain.FriendshipRequest ,error) {
	return s.friendsStore.CreateFriendRequest(ctx, params.CreateFriendRequestParams)
}

func (s *service) GetManyByUser(ctx context.Context, userId uuid.UUID) ([]domain.Friend, error) {
	return s.friendsStore.GetManyByUser(ctx, userId)
}

func (s *service) Create(ctx context.Context, params *store.CreateFriendParams) (domain.Friend, error) {
	return s.friendsStore.Create(ctx, params)
}

func (s *service) Delete(ctx context.Context, friendshipId uuid.UUID) error {
	return s.friendsStore.Delete(ctx, friendshipId)
}

func (s *service) AcceptFriendshipRequest(ctx context.Context, requestId uuid.UUID) (domain.Friendship, error) {
	return s.friendsStore.AcceptFrienshipRequest(ctx, requestId)
}

func (s *service) DeleteFriendshipRequest(ctx context.Context, id uuid.UUID) error {
	return s.friendsStore.DeleteFriendshipRequest(ctx, id)
}

func (s *service) GetFriendshipRequests(ctx context.Context, userId uuid.UUID) ([]domain.FriendshipRequest, error) {
	return s.friendsStore.GetFriendshipRequests(ctx, userId)
}
