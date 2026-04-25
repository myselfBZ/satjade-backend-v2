package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)
type practiceStore struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func (s *practiceStore) GetWithModules(ctx context.Context, id uuid.UUID) (*domain.Practice, error) {
	rows, err := s.queries.GetPracticeWithModules(ctx, pgtype.UUID{Bytes: id, Valid: true})

	if err != nil || len(rows) == 0 {
		if len(rows) == 0{
			return nil, domain.ErrRecordNotFound
		}

		return nil, err
	}

	practice := &domain.Practice{
		ID: rows[0].ID.String(),
		Title: rows[0].Title,
		CreatedAt: &rows[0].CreatedAt.Time,
		UpdatedAt: &rows[0].UpdatedAt.Time,
	}

	for i := range rows {
		practice.Modules = append(practice.Modules, domain.Module{
			ID: rows[i].ModuleID.String(),
			Name: rows[i].ModuleName,
			OrderIdx: rows[i].OrderIndex,
			Questions: []domain.ModuleQuestion{},
		})
	}

	return practice, nil
}



func (s *practiceStore) GetPreviews(ctx context.Context) ([]domain.Practice, error) {
	rows, err := s.queries.GetPracticePreviews(ctx)

	if err != nil {
		return nil, err
	}

	previews := make([]domain.Practice, len(rows))

	for i := 0; i < len(rows); i++ {
		previews[i] = domain.Practice{
			ID:        rows[i].ID.String(),
			Title:     rows[i].Title,
			CreatedAt: &rows[i].CreatedAt.Time,
			UpdatedAt: &rows[i].UpdatedAt.Time,

			Modules: nil,
		}
	}

	return previews, nil
}

func (s *practiceStore) CreateWithModules(ctx context.Context, title string) (*domain.Practice, error) {
	row, err := s.queries.CreatePracticeWithModules(ctx, title)

	if err != nil {
		return nil, err
	}

	modules, err := s.queries.GetModulesByPracticeID(ctx, row.ID)

	if err != nil {
		return nil, err
	}

	p := &domain.Practice{
		ID:        row.ID.String(),
		Title:     row.Title,
		CreatedAt: &row.CreatedAt.Time,
		UpdatedAt: &row.UpdatedAt.Time,
	}

	for i := 0; i < len(modules); i++ {
		p.Modules = append(p.Modules, domain.Module{
			ID: modules[i].ID.String(),
			Name: modules[i].Name,
			OrderIdx: modules[i].OrderIndex,
			Questions: nil,
		})
	}

	return 	p, nil
}
