package practiceattempt_service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/services/questions/answereval"
	"github.com/myselfBZ/satjade-backend/internal/store"
)

type gradeResult struct {
	math      int
	rw        int
	responses []domain.AttemptResponse
}

type CreatePracticeAttemptParams struct {
	PracticeId uuid.UUID `json:"practice_id" validate:"required"`
	UserId     uuid.UUID `json:"-"`
	StartedAt  time.Time `json:"started_at" validate:"required"`

	// question id -> choice_id or open answer key
	Responses map[string]string `json:"responses" validate:"required"`
}

type PracticeAttemptService interface {
	Create(ctx context.Context, params *CreatePracticeAttemptParams) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.PracticeAttempt, error)
	GetPreviewsByUser(ctx context.Context, userId uuid.UUID) ([]domain.PracticeAttemptPreview, error)
}

type ServiceParams struct {
	PracticeAttemptStore   store.PracticeAttemptStore
	PublishedPracticeStore store.PublishedPracticesStore
}

func New(params *ServiceParams) PracticeAttemptService {
	return &service{
		practiceAttemptStore:   params.PracticeAttemptStore,
		publishedPracticeStore: params.PublishedPracticeStore,
	}
}

type service struct {
	practiceAttemptStore   store.PracticeAttemptStore
	publishedPracticeStore store.PublishedPracticesStore
}

func (s *service) Create(ctx context.Context, params *CreatePracticeAttemptParams) error {
	p := &domain.PracticeAttempt{
		PracticeId:  params.PracticeId,
		UserId:      params.UserId,
		StartedAt:   params.StartedAt,
		SubmittedAt: time.Now(),
	}

	practice, err := s.publishedPracticeStore.GetById(ctx, params.PracticeId)

	if err != nil {
		return err
	}

	g, err := s.grade(practice.Data.Modules, params.Responses)

	if err != nil {
		return err
	}

	rwScaled, mathScaled, _ := Score(g.rw, g.math)

	p.RWScore = int16(rwScaled)
	p.MathScore = int16(mathScaled)
	p.Responses = g.responses

	return s.practiceAttemptStore.Create(ctx, p)
}

func (s *service) GetById(ctx context.Context, id uuid.UUID) (*domain.PracticeAttempt, error) {
	attempt, err := s.practiceAttemptStore.GetById(ctx, id)

	if err != nil {
		return nil, err
	}

	practice, err := s.publishedPracticeStore.GetById(ctx, attempt.PracticeId)

	if err != nil {
		return nil, err
	}

	responses := []domain.AttemptResponse{}

	for _, m := range practice.Data.Modules {
		for _, q := range m.Questions {
			found := false
			for _, r := range attempt.Responses {
				if r.QuestionId == q.ID {
					r.Question = &q
					responses = append(responses, r)
					found = true
				}
			}

			if !found {
				responses = append(responses, domain.AttemptResponse{
					Question: &q,
					IsCorrect: false,
				})
			}

		}
	}

	attempt.Responses = responses

	return attempt, nil
}

func (s *service) GetPreviewsByUser(ctx context.Context, userId uuid.UUID) ([]domain.PracticeAttemptPreview, error) {
	return s.practiceAttemptStore.GetPreviewsByUser(ctx, userId)
}

func (s *service) grade(modules []domain.Module, responses map[string]string) (gradeResult, error) {
	//correct answers
	math := 0
	rw := 0

	domainResponses := []domain.AttemptResponse{}

	for _, m := range modules {
		for _, q := range m.Questions {
			resp, ok := responses[q.ID.String()]
			if ok {
				r := domain.AttemptResponse{
					Question: &domain.ModuleQuestion{
						Question: domain.Question{
							ID: q.ID,
						},
					},
				}

				switch q.Type {
				case "multiple_choice":
					var err error
					r.SelectedChoiceId, err = findMatchingChoiceId(resp, q.Choices)

					if err != nil {
						return gradeResult{}, nil
					}

					r.IsCorrect = *r.SelectedChoiceId == *q.CorrectChoiceID
					domainResponses = append(domainResponses, r)
				case "open_response":
					r.UserAnswer = &resp
					r.IsCorrect, _ = answereval.EvaluateAnswer(resp, q.OpenAnswerKey.ModelAnswer)
					domainResponses = append(domainResponses, r)
				}

				if r.IsCorrect {
					switch m.Name {
					case "Reading And Writing 1", "Reading And Writing 2":
						rw++
					case "Math 1", "Math 2":
						math++
					}
				}
			}
		}

	}

	return gradeResult{
		math:      math,
		rw:        rw,
		responses: domainResponses,
	}, nil
}

const (
	RWTotal   = 54
	MathTotal = 44
)

func scaleToSection(rawCorrect int, sectionTotal int) int {
	if sectionTotal <= 0 {
		return 200
	}
	pct := float64(rawCorrect) / float64(sectionTotal)
	scaled := 200.0 + pct*600.0
	return roundToNearest10(scaled)
}

func Score(rawCorrectRW, rawCorrectMath int) (int, int, int) {
	if rawCorrectRW < 0 {
		rawCorrectRW = 0
	}
	if rawCorrectMath < 0 {
		rawCorrectMath = 0
	}
	if rawCorrectRW > RWTotal {
		rawCorrectRW = RWTotal
	}
	if rawCorrectMath > MathTotal {
		rawCorrectMath = MathTotal
	}

	scaledRW := scaleToSection(rawCorrectRW, RWTotal)
	scaledMath := scaleToSection(rawCorrectMath, MathTotal)
	total := scaledRW + scaledMath
	return scaledRW, scaledMath, total
}

func roundToNearest10(x float64) int {
	return int(math.Round(x/10.0) * 10)
}

func findMatchingChoiceId(id string, choices []domain.AnswerChoice) (*uuid.UUID, error) {
	for _, c := range choices {
		if c.ID.String() == id {
			return &c.ID, nil
		}
	}
	return nil, domain.ErrInvalidChoice
}
