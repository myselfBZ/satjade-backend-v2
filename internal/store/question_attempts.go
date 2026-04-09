package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

type questionAttemptsStore struct {
	queries *db.Queries
}

func (s *questionAttemptsStore) Create(ctx context.Context, qa *domain.QuestionAttempt) (*domain.QuestionAttempt, error) {
	result, err := s.queries.CreateQuestionAttempt(ctx, db.CreateQuestionAttemptParams{
		Response:       qa.Response,
		ElapsedSeconds: int32(qa.ElapsedSeconds),
		QuestionID:     pgtype.UUID{Bytes: qa.QuestionID, Valid: true},
		UserID:         pgtype.UUID{Bytes: qa.UserID, Valid: true},
		IsCorrect:      qa.IsCorrect,
	})
	if err != nil {
		return nil, err
	}

	qa.QuestionID = result.QuestionID.Bytes
	qa.UserID = result.UserID.Bytes
	qa.CreatedAt = result.CreatedAt.Time
	qa.ElapsedSeconds = int(result.ElapsedSeconds)
	return qa, nil
}

func (s *questionAttemptsStore) GetByUser(ctx context.Context, id uuid.UUID) ([]domain.QuestionAttempt, error) {
	results, err := s.queries.GetQuestionAttemptsByUser(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, err
	}

	attempts := make([]domain.QuestionAttempt, len(results))
	for i, r := range results {
		attempts[i] = domain.QuestionAttempt{
			Response:       r.Response,
			ElapsedSeconds: int(r.ElapsedSeconds),
			QuestionID:     r.QuestionID.Bytes,
			UserID:         r.UserID.Bytes,
			CreatedAt:      r.CreatedAt.Time,
			IsCorrect:      r.IsCorrect,
		}
	}
	return attempts, nil
}
