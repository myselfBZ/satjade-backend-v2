package practices_service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/store"
)

type service struct {
	practiceStore          store.PracticeStore
	questionStore          store.QuestionStore
	publishedPracticeStore store.PublishedPracticesStore
}

type CreatePracticeParams struct {
	Title string `json:"title" validate:"required"`
}

type ServiceParams struct {
	PracticeStore          store.PracticeStore
	QuestionStore          store.QuestionStore
	PublishedPracticeStore store.PublishedPracticesStore
}

func New(params *ServiceParams) PracticeService {
	return &service{
		practiceStore:          params.PracticeStore,
		questionStore:          params.QuestionStore,
		publishedPracticeStore: params.PublishedPracticeStore,
	}
}

type PracticeService interface {
	GetPublishedPreviews(ctx context.Context) ([]domain.PublishedPracticePreview, error)
	Create(ctx context.Context, title string) (*domain.Practice, error)
	GetPreviews(ctx context.Context) ([]domain.Practice, error)
	GetFullTest(ctx context.Context, id uuid.UUID) (*domain.Practice, error)
	Publish(ctx context.Context, params *PublishParams) error
	GetPublishedById(ctx context.Context, id uuid.UUID) (*domain.PublishedPractice, error)
}

type PublishParams struct {
	PublishedBy uuid.UUID
	PracticeId  uuid.UUID
}

func (s *service) GetPublishedPreviews(ctx context.Context) ([]domain.PublishedPracticePreview, error)  {
	return s.publishedPracticeStore.GetPreviews(ctx)
}

func (s *service) Publish(ctx context.Context, params *PublishParams) error {
	practice, err := s.GetFullTest(ctx, params.PracticeId)

	if err != nil {
		return err
	}

	practice.CreatedAt = nil
	practice.UpdatedAt = nil

	data, err := json.Marshal(practice)

	if err != nil {
		return err
	}

	return s.publishedPracticeStore.Publish(ctx, &store.PublishParams{
		PracticeID:  params.PracticeId,
		PublishedBy: params.PublishedBy,
		Data:        data,
	})
}

func (s *service) GetPublishedById(ctx context.Context, id uuid.UUID) (*domain.PublishedPractice, error) {
	publishedTest, err := s.publishedPracticeStore.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	for i := range publishedTest.Data.Modules {
		for qIdx := range publishedTest.Data.Modules[i].Questions {
			ptr := &publishedTest.Data.Modules[i].Questions[qIdx]
			ptr.Question.HideKeys()
		}
	}

	return publishedTest, nil
}

func (s *service) Create(ctx context.Context, title string) (*domain.Practice, error) {
	return s.practiceStore.CreateWithModules(ctx, title)
}

func (s *service) GetPreviews(ctx context.Context) ([]domain.Practice, error) {
	return s.practiceStore.GetPreviews(ctx)
}

func (s *service) GetFullTest(ctx context.Context, id uuid.UUID) (*domain.Practice, error) {
	practice, err := s.practiceStore.GetWithModules(ctx, id)

	if err != nil {
		return nil, err
	}

	for i := range practice.Modules {
		validUUID, _ := uuid.Parse(practice.Modules[i].ID)
		m := &practice.Modules[i]
		m.Questions, err = s.questionStore.GetByModule(ctx, validUUID)
		if err != nil {
			return nil, err
		}
	}

	return practice, nil
}
