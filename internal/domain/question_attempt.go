package domain

import (
	"time"

	"github.com/google/uuid"
)

type QuestionAttempt struct {
	ID             uuid.UUID `json:"id"`
	Response       string    `json:"response" validate:"required"`
	ElapsedSeconds int       `json:"elapsed_seconds" validate:"required"`
	QuestionID     uuid.UUID `json:"question_id" validate:"required"`
	UserID         uuid.UUID `json:"user_id" validate:"required"`
	CreatedAt      time.Time `json:"created_at"`
	IsCorrect      bool      `json:"is_correct"`
}

