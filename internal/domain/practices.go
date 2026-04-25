package domain

import (
	"time"

	"github.com/google/uuid"
)

type Practice struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Modules   []Module   `json:"modules,omitempty"`
}

type Module struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	OrderIdx  int16            `json:"order_idx"`
	Questions []ModuleQuestion `json:"questions,omitempty"`
}

type PublishParams struct {
	PublishedBy uuid.UUID
	PracticeID  uuid.UUID
	Data        []byte
}

type PublishedPractice struct {
	ID          string    `json:"id"`
	PublishedAt time.Time `json:"published_at"`
	Data        *Practice `json:"data"`
}

type PublishedPracticePreview struct {
	ID          uuid.UUID `json:"id"`
	Title 		string 	  `json:"title"`
	PublishedAt time.Time `json:"published_at"`
}
