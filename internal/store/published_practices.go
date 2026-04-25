package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

type publishedPracticeStore struct {
	queries *db.Queries
}

type PublishParams struct {
	PublishedBy uuid.UUID
	PracticeID  uuid.UUID
	Data        []byte
}

func (s *publishedPracticeStore) GetPreviews(ctx context.Context) ([]domain.PublishedPracticePreview, error) {
	rows, err := s.queries.GetPublishedPracticesPreviews(ctx)

	if err != nil {
		return nil, err
	}

	previews := make([]domain.PublishedPracticePreview, len(rows))

	for i := range rows {
		previews[i] = domain.PublishedPracticePreview{
			ID:          rows[i].ID.Bytes,
			Title:       rows[i].Title,
			PublishedAt: rows[i].PublishedAt.Time,
		}
	}

	return previews, nil
}

func (s *publishedPracticeStore) GetById(ctx context.Context, id uuid.UUID) (*domain.PublishedPractice, error) {
	row, err := s.queries.GetPublishedPractice(ctx, pgtype.UUID{Bytes: id, Valid: true})

	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, domain.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	var practice domain.Practice

	if err := json.Unmarshal(row.Data, &practice); err != nil {
		return nil, err
	}

	return &domain.PublishedPractice{
		ID:          row.ID.String(),
		PublishedAt: row.PublishedAt.Time,
		Data:        &practice,
	}, nil
}

func (s *publishedPracticeStore) Publish(ctx context.Context, params *PublishParams) error {
	_, err := s.queries.PublishPractice(ctx, db.PublishPracticeParams{
		PublishedBy: pgtype.UUID{Bytes: params.PublishedBy, Valid: true},
		PracticeID:  pgtype.UUID{Bytes: params.PracticeID, Valid: true},
		Data:        params.Data,
	})

	return err
}
