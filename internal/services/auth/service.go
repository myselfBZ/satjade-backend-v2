package auth_service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/myselfBZ/satjade-backend/internal/auth"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type AuthSuccessResponse struct {
	AccessToken string       `json:"access_token"`
	User        *domain.User `json:"user"`
}

type LoginParams struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type CreateAccountParams struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Email    string `json:"email" validate:"required,email,max=255"`
}

type ServiceParams struct {
	Authenticator auth.Authenticator
	UserStore     store.UserStore
	Issuer        string
	Aud           string
	ExpTime       time.Duration
}

func New(params *ServiceParams) AuthService {
	return &service{
		auth:      params.Authenticator,
		iss:       params.Issuer,
		aud:       params.Aud,
		expTime:   params.ExpTime,
		userStore: params.UserStore,
	}
}

type AuthService interface {
	Login(ctx context.Context, params *LoginParams) (*AuthSuccessResponse, error)
	CreateAccount(ctx context.Context, params *CreateAccountParams) (*AuthSuccessResponse, error)
}

type service struct {
	auth      auth.Authenticator
	iss       string
	aud       string
	userStore store.UserStore
	expTime   time.Duration
}

func (a *service) CreateAccount(ctx context.Context, params *CreateAccountParams) (*AuthSuccessResponse, error) {
	user := &domain.User{
		FullName: params.FullName,
		Email:    params.Email,
		Role:     domain.ROLE_STUDENT,
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user.Password = string(hash)

	if err := a.userStore.Create(ctx, user); err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(a.expTime).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": a.iss,
		"aud": a.aud,
	}

	token, err := a.auth.GenerateToken(claims)

	if err != nil {
		return nil, err
	}

	return &AuthSuccessResponse{
		AccessToken: token,
		User:        user,
	}, nil
}

func (a *service) Login(ctx context.Context, params *LoginParams) (*AuthSuccessResponse, error) {

	user, err := a.userStore.GetByEmail(ctx, params.Email)

	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(a.expTime).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": a.iss,
		"aud": a.aud,
	}

	token, err := a.auth.GenerateToken(claims)

	if err != nil {
		return nil, err
	}

	return &AuthSuccessResponse{
		AccessToken: token,
		User:        user,
	}, nil

}
