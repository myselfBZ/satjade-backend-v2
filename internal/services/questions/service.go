package questions_service

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/filestore"
	"github.com/myselfBZ/satjade-backend/internal/services/questions/answereval"
	"github.com/myselfBZ/satjade-backend/internal/store"
)

type CreateQuestParams struct {
	Type          string                `json:"type"`
	ImageBase64   string                `json:"image_base64"`
	Paragraph     *string               `json:"paragraph,omitempty"`
	Prompt        string                `json:"prompt"`
	Skill         string                `json:"skill"`
	Domain        string                `json:"domain"`
	Difficulty    string                `json:"difficulty"`
	Explanation   string                `json:"explanation"`
	OpenAnswerKey *domain.OpenAnswerKey `json:"open_answer_key,omitempty"`
	Choices       []domain.AnswerChoice `json:"choices,omitempty"`
}

type CreateAttemptParams struct {
	domain.QuestionAttempt
}

type CheckParams struct {
	Response   string
	QuestionId uuid.UUID
}

type CreateToModuleParams struct {
	Question *CreateQuestParams `json:"question"`
	Number   int                `json:"number"`
	// set by the handler manually
	ModuleId uuid.UUID `json:"-"`
}

type LinkToModuleParams struct {
	store.LinkToModuleParams
}

type FilterParams struct {
	store.FilterParams
}

type ServiceParams struct {
	QuestionStore         store.QuestionStore
	QuestionAttemptsStore store.QuestionAttemptsStore
	FileStore             filestore.FileStorage
}

func New(params *ServiceParams) QuestionsService {
	return &service{
		questionStore:         params.QuestionStore,
		questionAttemptsStore: params.QuestionAttemptsStore,
		fileStore:             params.FileStore,
	}
}

type QuestionsService interface {
	Create(ctx context.Context, params *CreateQuestParams) (*domain.Question, error)
	GetById(ctx context.Context, id uuid.UUID) (*domain.Question, error)
	GetDistribution(ctx context.Context, userId uuid.UUID) ([]domain.DomainDistribution, error)
	CreateToModule(ctx context.Context, params *CreateToModuleParams) error
	LinkToModule(ctx context.Context, params *LinkToModuleParams) error
	FilterIDs(ctx context.Context, params *FilterParams) (uuid.UUIDs, error)
	GetAttemptsByUser(ctx context.Context, userID uuid.UUID) ([]domain.QuestionAttempt, error)
	CreateAttempt(ctx context.Context, params *CreateAttemptParams) (*domain.QuestionAttempt, error)
	GetRandomIds(ctx context.Context) (uuid.UUIDs, error)
	Check(ctx context.Context, params CheckParams) (bool ,error)
}

type service struct {
	questionStore         store.QuestionStore
	questionAttemptsStore store.QuestionAttemptsStore
	fileStore             filestore.FileStorage
}

// TODO make it an atomic operation
func (s *service) CreateToModule(ctx context.Context, params *CreateToModuleParams) error {
	q, err := s.Create(ctx, params.Question)

	if err != nil {
		return err
	}

	return s.LinkToModule(ctx, &LinkToModuleParams{
		LinkToModuleParams: store.LinkToModuleParams{
			Number:     params.Number,
			ModuleId:   params.ModuleId,
			QuestionId: q.ID,
		},
	})
}

func (s *service) LinkToModule(ctx context.Context, params *LinkToModuleParams) error {
	return s.questionStore.LinkToModule(ctx, &params.LinkToModuleParams)
}

func (s *service) Create(ctx context.Context, params *CreateQuestParams) (*domain.Question, error) {
	question := &domain.Question{
		Type:        params.Type,
		Prompt:      params.Prompt,
		Explanation: params.Explanation,
		Skill:       params.Skill,
		Domain:      params.Domain,
		Difficulty:  params.Difficulty,
	}

	if params.Paragraph != nil {
		question.Paragraph = params.Paragraph
	}

	if params.Type == "multiple_choice" {
		question.Choices = params.Choices
	} else {
		question.OpenAnswerKey = params.OpenAnswerKey
	}

	var filename string

	if params.ImageBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(params.ImageBase64)

		if err != nil {
			return nil, err
		}

		reader := bytes.NewReader(decoded)

		filename, err = s.fileStore.Save("image.jpg", reader)

		if err != nil {
			return nil, err
		}
	}

	question.ImagePath = &filename

	if err := s.questionStore.Create(ctx, question); err != nil {
		s.fileStore.Delete(filename)
		return nil, err
	}

	return question, nil
}

func (s *service) GetById(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	return s.questionStore.GetById(ctx, id)
}

func (s *service) GetDistribution(ctx context.Context, userId uuid.UUID) ([]domain.DomainDistribution, error) {
	return s.questionStore.GetDistributionByDomain(ctx, userId)
}

func (s *service) FilterIDs(ctx context.Context, params *FilterParams) (uuid.UUIDs, error) {

	if len(params.Domains) == 0 {
		params.Domains = []string{"all"}
	}

	if len(params.Skills) == 0 {
		params.Skills = []string{"all"}
	}

	if len(params.Difficulty) == 0 {
		params.Difficulty = []string{"easy", "medium", "hard"}
	}

	if params.AttemptStatus == "" {
		params.AttemptStatus = "all"
	}

	return s.questionStore.FilterIDs(ctx, &params.FilterParams)
}

func (s *service) GetAttemptsByUser(ctx context.Context, userID uuid.UUID) ([]domain.QuestionAttempt, error) {
	return s.questionAttemptsStore.GetByUser(ctx, userID)
}

func (s *service) CreateAttempt(ctx context.Context, params *CreateAttemptParams) (*domain.QuestionAttempt, error) {
	question, err := s.questionStore.GetById(ctx, params.QuestionID)

	if err != nil {
		return nil, err
	}

	if question.Type == "multiple_choice" {
		upper := strings.ToUpper(params.Response)
		params.IsCorrect = upper == store.MustGetCorrectChoice(question.Choices).Label
	}

	if question.Type == "open_response" {
		correct, err := answereval.EvaluateAnswer(params.Response, question.OpenAnswerKey.ModelAnswer)

		if err != nil || !correct {
			params.IsCorrect = false
		} else {
			params.IsCorrect = true
		}

	}

	return s.questionAttemptsStore.Create(ctx, &params.QuestionAttempt)
}

func (s *service) GetRandomIds(ctx context.Context) (uuid.UUIDs, error) {
	return s.questionStore.GetRandomIds(ctx)
}

func (s *service) Check(ctx context.Context, params CheckParams) (bool ,error) {
	isCorrect := false
	question, err := s.questionStore.GetById(ctx, params.QuestionId)

	if err != nil {
		return false, err
	}

	if question.Type == "multiple_choice" {
		upper := strings.ToUpper(params.Response)
		isCorrect = upper == store.MustGetCorrectChoice(question.Choices).Label
	}

	if question.Type == "open_response" {
		correct, err := answereval.EvaluateAnswer(params.Response, question.OpenAnswerKey.ModelAnswer)

		if err != nil || !correct {
			isCorrect = false
		} else {
			isCorrect = true
		}

	}

	return isCorrect, nil
}




