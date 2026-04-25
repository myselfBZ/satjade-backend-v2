package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/myselfBZ/satjade-backend/internal/db"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

type userStore struct {
	queries *db.Queries
}

func (s *userStore) Create(ctx context.Context, u *domain.User) error {
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		FullName:     u.FullName,
		Email:        u.Email,
		PasswordHash: u.Password,
		Role:         u.Role,
	})

	if err != nil {
		switch {
		case err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`:
			return domain.ErrDuplicateEmail
		default:
			return err
		}
	}

	u.ID = user.ID.Bytes
	return nil
}

func (s *userStore) GetMany(ctx context.Context) ([]domain.User, error) {
	userRows, err := s.queries.GetUsers(ctx)
	var users []domain.User
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users, nil
		}
		return nil, err
	}

	for _, u := range userRows {
		users = append(users, domain.User{
			ID:        u.ID.Bytes,
			Email:     u.Email,
			FullName:  u.FullName,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Time,
			UpdatedAt: u.UpdatedAt.Time,
		})
	}

	return users, nil
}

func (s *userStore) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.queries.DeleteUser(ctx, pgtype.UUID{
		Bytes: id,
		Valid: true,
	})

	return err
}

func (s *userStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.queries.GetUserById(ctx, pgtype.UUID{Bytes: id, Valid: true})

	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, domain.ErrRecordNotFound
		default:
			return nil, err
		}

	}

	return &domain.User{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		Password:  user.PasswordHash,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *userStore) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, domain.ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &domain.User{
		ID:        user.ID.Bytes,
		FullName:  user.FullName,
		Email:     user.Email,
		Password:  user.PasswordHash,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}
