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
	}
}

type Storage struct {
	Users              UserStore
	Practices          PracticeStore
	Questions          QuestionStore
	QuestionAttempts   QuestionAttemptsStore
	PublishedPractices PublishedPracticesStore
}

type QuestionStore interface {
	Create(ctx context.Context, q *domain.Question) error
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

func deref(s *string) string {
	if s == nil {
		return ""
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
