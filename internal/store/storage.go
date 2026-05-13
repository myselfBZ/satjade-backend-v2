package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

func New(pool *pgxpool.Pool) *Storage {
	queries := db.New(pool)

	return &Storage{
		Users:              &userStore{queries},
		Practices:          &practiceStore{queries: queries, pool: pool},
		Questions:          &questionStore{queries: *queries, pool: pool},
		PublishedPractices: &publishedPracticeStore{queries: queries},
		QuestionAttempts:   &questionAttemptsStore{queries: queries},
		PracticeAttempts:   &practiceAttemptsStore{queries: queries, pool: pool},
		Friends:            &friendsStore{queries: queries, pool: pool},
	}
}

type Storage struct {
	Users              UserStore
	Practices          PracticeStore
	Questions          QuestionStore
	QuestionAttempts   QuestionAttemptsStore
	PublishedPractices PublishedPracticesStore
	PracticeAttempts   PracticeAttemptStore
	Friends            FriendsStore
}

type PracticeAttemptStore interface {
	Create(ctx context.Context, p *domain.PracticeAttempt) error
	GetPreviewsByUser(ctx context.Context, userId uuid.UUID) ([]domain.PracticeAttemptPreview, error)
	GetById(ctx context.Context, id uuid.UUID) (*domain.PracticeAttempt, error)
}

type QuestionStore interface {
	Create(ctx context.Context, q *domain.Question) error
	GetRandomIds(ctx context.Context) (uuid.UUIDs, error) 
	GetById(ctx context.Context, id uuid.UUID) (*domain.Question, error)
	GetDistributionByDomain(ctx context.Context, userId uuid.UUID) ([]domain.DomainDistribution, error)
	CreateToModule(ctx context.Context, params *CreateToModuleParam) error
	LinkToModule(ctx context.Context, params *LinkToModuleParams) error
	GetByModule(ctx context.Context, moduleId uuid.UUID) ([]domain.ModuleQuestion, error)
	FilterIDs(ctx context.Context, params *FilterParams) (uuid.UUIDs, error)
}

type QuestionAttemptsStore interface {
	Create(ctx context.Context, qa *domain.QuestionAttempt) (*domain.QuestionAttempt, error)
	GetByUser(ctx context.Context, id uuid.UUID) ([]domain.QuestionAttempt, error)
}

type PublishedPracticesStore interface {
	Publish(ctx context.Context, params *PublishParams) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.PublishedPractice, error)
	GetPreviews(ctx context.Context) ([]domain.PublishedPracticePreview, error)
}

type PracticeStore interface {
	CreateWithModules(ctx context.Context, title string) (*domain.Practice, error)
	GetPreviews(ctx context.Context) ([]domain.Practice, error)
	GetWithModules(ctx context.Context, id uuid.UUID) (*domain.Practice, error)
}

type UserStore interface {
	Create(ctx context.Context, u *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetMany(ctx context.Context) ([]domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type FriendsStore interface {
	GetManyByUser(ctx context.Context, userId uuid.UUID) ([]domain.Friend, error)
	Create(ctx context.Context, params *CreateFriendParams) (domain.Friend, error)
	Delete(ctx context.Context, friendshipId uuid.UUID) error
	CreateFriendRequest(ctx context.Context, params *CreateFriendRequestParams) (domain.FriendshipRequest ,error)
	DeleteFriendshipRequest(ctx context.Context, id uuid.UUID) error
	AcceptFrienshipRequest(ctx context.Context, requestId uuid.UUID) (domain.Friendship, error)
	GetFriendshipRequests(ctx context.Context, userId uuid.UUID) ([]domain.FriendshipRequest, error)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefUUID(s *uuid.UUID) uuid.UUID {
	if s == nil {
		return uuid.UUID{}
	}
	return *s
}

func withTx(pool *pgxpool.Pool, ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
