package domain

import "github.com/google/uuid"

type Question struct {
	ID              uuid.UUID      `json:"id"`
	Type            string         `json:"type"`
	Paragraph       *string        `json:"paragraph,omitempty"`
	Prompt          string         `json:"prompt"`
	ImagePath       *string        `json:"image_path,omitempty"`
	Skill           string         `json:"skill"`
	Domain          string         `json:"domain"`
	Difficulty      string         `json:"difficulty"`
	Explanation     string         `json:"explanation"`
	CorrectChoiceID *uuid.UUID        `json:"correct_choice_id,omitempty"`
	OpenAnswerKey   *OpenAnswerKey `json:"open_answer_key,omitempty"`
	Choices         []AnswerChoice `json:"choices,omitempty"`
}

func (q *Question) HideKeys() {
	q.OpenAnswerKey = nil
	q.CorrectChoiceID = nil

	for chIdx := range q.Choices {
		q.Choices[chIdx].IsCorrect = false
	}

	q.Explanation = ""
}

type AnswerChoice struct {
	ID         uuid.UUID `json:"id"`
	QuestionID string    `json:"question_id,omitempty"`
	Label      string    `json:"label"`
	Body       string    `json:"body"`
	IsCorrect  bool      `json:"is_correct"`
}

type OpenAnswerKey struct {
	ID           uuid.UUID `json:"id"`
	QuestionID   string `json:"question_id,omitempty"`
	ModelAnswer  string `json:"model_answer"`
	MatchPattern string `json:"match_pattern"`
}

type DifficultyDistribution struct {
	Name        string `json:"name"`
	Count       int64  `json:"count"`
	Correct     int64  `json:"correct"`
	Incorrect   int64  `json:"incorrect"`
	Unattempted int64  `json:"unattempted"`
}

type SkillDistribution struct {
	Count      int                      `json:"count"`
	Name       string                   `json:"name"`
	Diffulties []DifficultyDistribution `json:"difficulties"`
}

type DomainDistribution struct {
	Domain     string `json:"domain"`
	TotalCount int64  `json:"total_count"`

	Skills []SkillDistribution `json:"skills"`
}

type ModuleQuestion struct {
	Number int `json:"number"`
	Question
}

