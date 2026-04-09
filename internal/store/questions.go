package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

type FilterParams struct {
	UserID        uuid.UUID `query:"-"`
	Domains       []string  `query:"domain"`
	Skills        []string  `query:"skill"`
	Difficulty    []string  `query:"difficulty"`
	AttemptStatus string    `query:"attempt_status"`
}

type CreateToModuleParam struct {
	Number int `json:"number" validate:"required"`
	// always set manually
	ModuleId uuid.UUID `json:"-" validate:"required"`
	Question *domain.Question `json:"question_id" validate:"required"`
}

type LinkToModuleParams struct {
	Number int `json:"number" validate:"required"`
	// always set manually
	ModuleId   uuid.UUID `json:"-" validate:"required,uuid"`
	QuestionId uuid.UUID `json:"question_id" validate:"required,uuid"`
}

type questionStore struct {
	queries db.Queries
	pool    *pgxpool.Pool
}

func (s *questionStore) GetByModule(ctx context.Context, moduleId uuid.UUID) ([]domain.ModuleQuestion, error) {
	rows, err := s.queries.GetQuestionsByModule(ctx, pgtype.UUID{Bytes: moduleId, Valid: true})
	if err != nil {
		return nil, err
	}

	questions := make([]domain.ModuleQuestion, 0, len(rows))

	for _, r := range rows {
		q := domain.ModuleQuestion{
			Number: int(r.Number),
			Question: domain.Question{
				ID:          r.ID.Bytes,
				Type:        string(r.Type),
				Prompt:      r.Prompt,
				Domain:      r.Domain.String,
				Difficulty:  string(r.Difficulty),
				Explanation: r.Explanation.String,
				Skill:       r.Skill.String,
			},
		}

		if r.Paragraph.Valid {
			q.Paragraph = &r.Paragraph.String
		}
		if r.ImagePath.Valid {
			q.ImagePath = &r.ImagePath.String
		}

		if len(r.Choices) > 0 {
			id := uuid.UUID(r.CorrectChoiceID.Bytes)
			q.CorrectChoiceID = &id
			if err := json.Unmarshal(r.Choices, &q.Choices); err != nil {
				return nil, fmt.Errorf("failed to unmarshal choices for question %s: %w", r.ID, err)
			}

			for i := range q.Choices {
				if q.Choices[i].ID == *q.CorrectChoiceID {
					q.Choices[i].IsCorrect = true
				}
			}
		}

		if r.OpenKey != nil {
			var openKey domain.OpenAnswerKey
			if err := json.Unmarshal(r.OpenKey, &openKey); err != nil {
				return nil, fmt.Errorf("failed to unmarshal open key for question %s: %w", r.ID, err)
			}
			if openKey.ID.String() != "" {
				q.OpenAnswerKey = &openKey
			}
		}

		questions = append(questions, q)
	}

	return questions, nil
}

func (s *questionStore) GetDistributionByDomain(ctx context.Context, userId uuid.UUID) ([]domain.DomainDistribution, error) {
	rows, err := s.queries.GetQuestionDistribution(ctx, pgtype.UUID{Bytes: userId, Valid: true})
	if err != nil {
		return nil, err
	}

	type domainTracker struct {
		totalCount int64
		skills     map[string][]domain.DifficultyDistribution
	}

	groups := make(map[string]*domainTracker)

	for _, row := range rows {
		dName := row.Domain.String
		sName := row.Skill.String

		if _, ok := groups[dName]; !ok {
			groups[dName] = &domainTracker{
				skills: make(map[string][]domain.DifficultyDistribution),
			}
		}

		groups[dName].totalCount += row.TotalCount

		groups[dName].skills[sName] = append(groups[dName].skills[sName], domain.DifficultyDistribution{
			Name:        row.Difficulty,
			Count:       row.TotalCount,
			Correct:     row.CorrectCount,
			Incorrect:   row.IncorrectCount,
			Unattempted: row.UnattemptedCount,
		})
	}

	result := make([]domain.DomainDistribution, 0, len(groups))
	for dName, data := range groups {
		domainRes := domain.DomainDistribution{
			Domain:     dName,
			TotalCount: data.totalCount,
			Skills:     make([]domain.SkillDistribution, 0, len(data.skills)),
		}

		for sName, diffs := range data.skills {
			count := 0
			for _, d := range diffs {
				count += int(d.Count)
			}

			domainRes.Skills = append(domainRes.Skills, domain.SkillDistribution{
				Name:       sName,
				Count: count,
				Diffulties: diffs,
			})
		}
		result = append(result, domainRes)
	}

	return result, nil
}

func (s *questionStore) CreateToModule(ctx context.Context, params *CreateToModuleParam) error {
	return withTx(s.pool, ctx, func(tx pgx.Tx) error {
		txConn := s.queries.WithTx(tx)
		if err := s.create(ctx, txConn, params.Question); err != nil {
			return err
		}

		if err := s.queries.LinkQuestionToModule(ctx, db.LinkQuestionToModuleParams{
			ModuleID:   pgtype.UUID{Bytes: params.ModuleId, Valid: true},
			QuestionID: pgtype.UUID{Bytes: params.Question.ID, Valid: true},
			Number:     int16(params.Number),
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *questionStore) LinkToModule(ctx context.Context, params *LinkToModuleParams) error {
	err := s.queries.LinkQuestionToModule(ctx, db.LinkQuestionToModuleParams{
		ModuleID:   pgtype.UUID{Bytes: params.ModuleId, Valid: true},
		QuestionID: pgtype.UUID{Bytes: params.QuestionId, Valid: true},
		Number:     int16(params.Number),
	})

	if err != nil {
		if strings.Contains(err.Error(), "module_questions_module_id_fkey") {
			return domain.ErrModuleNotFound
		}

		if strings.Contains(err.Error(), "module_questions_question_id_fkey") {
			return domain.ErrQuestionNotFound
		}

		return err
	}

	return nil
}

func (s *questionStore) create(ctx context.Context, tx *db.Queries, q *domain.Question) error {
	row, err := tx.CreateQuestion(ctx, db.CreateQuestionParams{
		Type:        db.QuestionType(q.Type),
		Paragraph:   pgtype.Text{String: deref(q.Paragraph), Valid: q.Paragraph != nil},
		Prompt:      q.Prompt,
		ImagePath:   pgtype.Text{String: deref(q.ImagePath), Valid: q.ImagePath != nil},
		Skill:       pgtype.Text{String: q.Skill, Valid: true},
		Domain:      pgtype.Text{String: q.Domain, Valid: true},
		Difficulty:  q.Difficulty,
		Explanation: pgtype.Text{String: q.Explanation, Valid: true},
	})

	if err != nil {
		return err
	}

	q.ID = row.ID.Bytes

	if q.Type == string(db.QuestionTypeMultipleChoice) {
		choices, err := s.createAnswerChoices(ctx, tx, row.ID, q.Choices)

		if err != nil {
			return err
		}

		correctCh, err := GetCorrectChoice(choices)

		if err != nil {
		}

		err = tx.SetCorrectChoice(ctx, db.SetCorrectChoiceParams{
			ID:              row.ID,
			CorrectChoiceID: pgtype.UUID{Bytes: correctCh.ID, Valid: true},
		})

		if err != nil {
			return err
		}
	} else {
		_, err := tx.CreateOpenAnswerKey(ctx, db.CreateOpenAnswerKeyParams{
			QuestionID:   row.ID,
			MatchPattern: pgtype.Text{String: q.OpenAnswerKey.MatchPattern, Valid: true},
			ModelAnswer:  pgtype.Text{String: q.OpenAnswerKey.ModelAnswer, Valid: true},
		})

		if err != nil {
			return err
		}
	}

	return nil

}

func (s *questionStore) Create(ctx context.Context, q *domain.Question) error {
	return withTx(s.pool, ctx, func(txConn pgx.Tx) error {
		tx := s.queries.WithTx(txConn)
		return s.create(ctx, tx, q)
	})
}

func (s *questionStore) GetById(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	row, err := s.queries.GetQuestionByID(ctx, pgtype.UUID{Bytes: id, Valid: true})

	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, domain.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	q := &domain.Question{
		Type:        string(row.Type),
		ID:          row.ID.Bytes,
		Prompt:      row.Prompt,
		Domain:      row.Domain.String,
		Difficulty:  string(row.Difficulty),
		Explanation: row.Explanation.String,
		Skill:       row.Skill.String,
	}

	if row.Paragraph.Valid {
		q.Paragraph = &row.Paragraph.String
	}

	if row.ImagePath.Valid {
		q.ImagePath = &row.ImagePath.String
	}

	if row.Type == db.QuestionTypeMultipleChoice {
		if err := json.Unmarshal(row.Choices, &q.Choices); err != nil {
			return nil, err
		}

		id := uuid.UUID(row.CorrectChoiceID.Bytes)

		q.CorrectChoiceID = &id

		markCorrectChoice(*q.CorrectChoiceID, q.Choices)
	} else {
		q.Choices = nil
		q.OpenAnswerKey = &domain.OpenAnswerKey{
			ID:           row.AnswerKeyID.Bytes,
			ModelAnswer:  row.AnswerKeyModelAnswer.String,
			MatchPattern: row.AnswerKeyMatchPattern.String,
		}
	}

	return q, nil
}

func GetCorrectChoice(choices []domain.AnswerChoice) (*domain.AnswerChoice, error) {
	for i := 0; i < len(choices); i++ {
		if choices[i].IsCorrect {
			return &choices[i], nil
		}
	}
	return nil, domain.ErrNoCorrectChoice
}

func MustGetCorrectChoice(choices []domain.AnswerChoice) *domain.AnswerChoice {
	for i := 0; i < len(choices); i++ {
		if choices[i].IsCorrect {
			return &choices[i]
		}
	}
	panic("MustGetCorrectChoice failed: no correct choices")
}

func (s *questionStore) createAnswerChoices(ctx context.Context, tx *db.Queries, questioId pgtype.UUID, choices []domain.AnswerChoice) ([]domain.AnswerChoice, error) {
	if len(choices) != 4 {
		return nil, domain.ErrInvalidNumberOfChoice
	}

	rows, err := tx.CreateAnswerChoices(ctx, db.CreateAnswerChoicesParams{
		QuestionID: questioId,
		Label:      choices[0].Label,
		Body:       choices[0].Body,
		Label_2:    choices[1].Label,
		Body_2:     choices[1].Body,
		Label_3:    choices[2].Label,
		Body_3:     choices[2].Body,
		Label_4:    choices[3].Label,
		Body_4:     choices[3].Body,
	})

	if err != nil {
		return nil, err
	}

	for _, r := range rows {
		for i := 0; i < len(choices); i++ {
			ptr := &choices[i]

			if ptr.Label == r.Label {
				ptr.ID = r.ID.Bytes
			}
		}
	}

	return choices, nil
}

func (s *questionStore) FilterIDs(ctx context.Context, params *FilterParams) (uuid.UUIDs, error) {
	rows, err := s.queries.FilterQuestions(ctx, db.FilterQuestionsParams{
		UserID:           pgtype.UUID{Bytes: params.UserID, Valid: true},
		DifficultyLevels: params.Difficulty,
		Domains:          params.Domains,
		Skills:           params.Skills,
		AttemptStatus:    params.AttemptStatus,
	})

	if err != nil {
		return nil, err
	}

	ids := make(uuid.UUIDs, len(rows))

	for i, r := range rows {
		ids[i] = r.Bytes
	}
	return ids, nil
}

func markCorrectChoice(correctID uuid.UUID, choices []domain.AnswerChoice) {
	for i := range choices {
		if choices[i].ID == correctID {
			choices[i].IsCorrect = true
			return
		}
	}
}
