package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

type practiceAttemptsStore struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func (s *practiceAttemptsStore) Create(ctx context.Context, p *domain.PracticeAttempt) error {
	return withTx(s.pool, ctx, func(tx pgx.Tx) error {
		txQueries := s.queries.WithTx(tx)

		if err := s.create(ctx, txQueries, p); err != nil {
			return err
		}

		if err := s.createResponses(ctx, txQueries, p.ID, p.Responses); err != nil {
			return err
		}

		return nil
	})
}

func (s *practiceAttemptsStore) createResponses(ctx context.Context, tx *db.Queries, attemptId uuid.UUID, responses []domain.AttemptResponse) error {
	params := db.CreateAttemptResponsesParams{
		Column1: make([]pgtype.UUID, len(responses)),
		Column2: make([]pgtype.UUID, len(responses)),
		Column3: make([]pgtype.UUID, len(responses)),
		Column4: make([]string, len(responses)),
		Column5: make([]bool, len(responses)),
	}

	for i, r := range responses {
		params.Column1[i] = pgtype.UUID{Bytes: attemptId, Valid: true}
		params.Column2[i] = pgtype.UUID{Bytes: r.Question.ID, Valid: true}
		params.Column3[i] = pgtype.UUID{Bytes: derefUUID(r.SelectedChoiceId), Valid: r.SelectedChoiceId != nil}
		params.Column4[i] = deref(r.UserAnswer)
		params.Column5[i] = r.IsCorrect
	}

	return tx.CreateAttemptResponses(ctx, params)
}

func (s *practiceAttemptsStore) create(ctx context.Context, tx *db.Queries, p *domain.PracticeAttempt) error {
	r, err := tx.CreatePracticeAttempt(ctx, db.CreatePracticeAttemptParams{
		MathScore: p.MathScore,
		RwScore:   p.RWScore,
		UserID: pgtype.UUID{
			Bytes: p.UserId,
			Valid: true,
		},
		PublishedPracticeID: pgtype.UUID{
			Bytes: p.PracticeId,
			Valid: true,
		},
		StartedAt: pgtype.Timestamptz{
			Time:  p.StartedAt,
			Valid: true,
		},
		SubmittedAt: pgtype.Timestamptz{
			Time:  p.SubmittedAt,
			Valid: true,
		},
	})

	if err != nil {
		return err
	}

	p.ID = r.ID.Bytes

	return nil
}

func (s *practiceAttemptsStore) GetPreviewsByUser(ctx context.Context, userId uuid.UUID) ([]domain.PracticeAttemptPreview, error) {
	rows, err := s.queries.GetPracticeAttemptsByUser(ctx, pgtype.UUID{Bytes: userId, Valid: true})
	if err != nil {
		return nil, err
	}

	previews := make([]domain.PracticeAttemptPreview, len(rows))

	for i := range rows {
		previews[i] = domain.PracticeAttemptPreview{
			ID:          rows[i].ID.Bytes,
			PracticeId:  rows[i].PublishedPracticeID.Bytes,
			StartedAt:   rows[i].StartedAt.Time,
			SubmittedAt: rows[i].SubmittedAt.Time,
			RWScore:     rows[i].RwScore,
			MathScore:   rows[i].MathScore,
		}
	}

	return previews, nil
}

func (s *practiceAttemptsStore) GetById(ctx context.Context, id uuid.UUID) (*domain.PracticeAttempt, error) {
	row, err := s.queries.GetById(ctx, pgtype.UUID{Bytes: id, Valid: true})

	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, domain.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	p := &domain.PracticeAttempt{
		ID:          row.ID.Bytes,
		PracticeId:  row.PublishedPracticeID.Bytes,
		StartedAt:   row.StartedAt.Time,
		SubmittedAt: row.SubmittedAt.Time,
		RWScore:     row.RwScore,
		MathScore:   row.MathScore,
	}

	rows, err := s.queries.GetAttemptResponsesByAttempt(ctx, pgtype.UUID{Bytes: p.ID, Valid: true})

	if err != nil {
		return nil, err
	}

	responses := make([]domain.AttemptResponse, len(rows))

	for i := range rows {
		responses[i] = domain.AttemptResponse{
			QuestionId:        rows[i].QuestionID.Bytes,
			PracticeAttemptId: p.ID,
			UserAnswer:        &rows[i].OpenResponse.String,
			IsCorrect:         rows[i].IsCorrect.Bool,
		}

		if rows[i].SelectedChoiceID.Valid {
			id := uuid.UUID(rows[i].SelectedChoiceID.Bytes)
			responses[i].SelectedChoiceId = &id
		}



		if responses[i].SelectedChoiceId != nil {
			responses[i].UserAnswer = nil
		}

		if responses[i].UserAnswer != nil {
			responses[i].SelectedChoiceId = nil
		}
	}

	p.Responses = responses

	return p, nil
}
