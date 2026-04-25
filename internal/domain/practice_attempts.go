package domain

import (
	"time"

	"github.com/google/uuid"
)

type PracticeAttemptPreview struct {
	ID          uuid.UUID `json:"id"`
	PracticeId  uuid.UUID `json:"practice_id"`
	UserId      uuid.UUID `json:"user_id"`
	StartedAt   time.Time `json:"started_at"`
	SubmittedAt time.Time `json:"submitted_at"`
	RWScore     int16     `json:"rw_score"`
	MathScore   int16     `json:"math_score"`
}

type PracticeAttempt struct {
	ID          uuid.UUID         `json:"id"`
	PracticeId  uuid.UUID         `json:"practice_id"`
	UserId      uuid.UUID         `json:"user_id"`
	StartedAt   time.Time         `json:"started_at"`
	SubmittedAt time.Time         `json:"submitted_at"`
	RWScore     int16             `json:"rw_score"`
	MathScore   int16             `json:"math_score"`
	Responses   []AttemptResponse `json:"responses"`
}

type AttemptResponse struct {
	Question *ModuleQuestion `json:"question"`

	QuestionId 		  uuid.UUID  `json:"-"`
	PracticeAttemptId uuid.UUID  `json:"-"`

	UserAnswer        *string    `json:"open_answer"`
	SelectedChoiceId  *uuid.UUID `json:"selected_choice_id"`
	IsCorrect         bool       `json:"is_correct"`
}
