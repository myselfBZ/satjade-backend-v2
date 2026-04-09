package domain

import (
	"github.com/google/uuid"
	"time"
)


const (
	ROLE_STUDENT = "student"
	ROLE_ADMIN   = "admin"
	ROLE_TUTOR   = "tutor"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
