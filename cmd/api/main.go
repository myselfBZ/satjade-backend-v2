package main

import (
	"log"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/myselfBZ/satjade-backend/internal/auth"
	"github.com/myselfBZ/satjade-backend/internal/filestore"
	"github.com/myselfBZ/satjade-backend/internal/postgres"
	auth_service "github.com/myselfBZ/satjade-backend/internal/services/auth"
	friends_service "github.com/myselfBZ/satjade-backend/internal/services/friends"
	practiceattempt_service "github.com/myselfBZ/satjade-backend/internal/services/practice-attempt"
	practices_service "github.com/myselfBZ/satjade-backend/internal/services/practices"
	questions_service "github.com/myselfBZ/satjade-backend/internal/services/questions"
	users_service "github.com/myselfBZ/satjade-backend/internal/services/users"
	"github.com/myselfBZ/satjade-backend/internal/store"
	"go.uber.org/zap"
)

func main() {
	cfg := &config{}
	cfg.Load()

	a := &api{
		wsClients: newWsClientsMap(),
		challenges: newChallengeMap(),
		duels: newDuelMap(),
		config: cfg,
		// TODO, make it configurable
		wsConnCloseCh: make(chan string, 100),
		eventCh: make(chan eventWrapper, 100),
	}



	logger := zap.Must(zap.NewProduction(zap.AddCaller())).Sugar()
	defer logger.Sync()
	a.logger = logger
	a.validator = validator.New()

	db, err := postgres.New(postgres.Config{
		Addr:        a.config.dbUrl,
		MaxConns:    15,
		MinConns:    15,
		MaxIdleTime: "15m",
	})

	if err != nil {
		panic(err)
	}

	a.filestorage, err = filestore.NewLocalFileStorage("media")

	if err != nil {
		panic(err)
	}

	storage := store.New(db)

	authenticator := auth.NewJWTAuthenticator(a.config.secretKey, a.config.auth.aud, a.config.auth.iss)

	a.auth = authenticator

	authservice := auth_service.New(&auth_service.ServiceParams{
		Authenticator: authenticator,
		UserStore:     storage.Users,
		Issuer:        a.config.auth.iss,
		Aud:           a.config.auth.aud,
		ExpTime:       a.config.auth.exp,
	})

	userservice := users_service.New(storage.Users)

	practicesservice := practices_service.New(&practices_service.ServiceParams{
		PracticeStore:          storage.Practices,
		QuestionStore:          storage.Questions,
		PublishedPracticeStore: storage.PublishedPractices,
	})

	questionsservice := questions_service.New(&questions_service.ServiceParams{
		QuestionStore:         storage.Questions,
		QuestionAttemptsStore: storage.QuestionAttempts,
		FileStore:             a.filestorage,
	})
	practiceattemptservice := practiceattempt_service.New(&practiceattempt_service.ServiceParams{
		PracticeAttemptStore:   storage.PracticeAttempts,
		PublishedPracticeStore: storage.PublishedPractices,
	})
	friendsservice := friends_service.New(storage.Friends)

	services := services{
		Auth:            authservice,
		Users:           userservice,
		Practices:       practicesservice,
		Questions:       questionsservice,
		PracticeAttempt: practiceattemptservice,
		Friends: friendsservice,
	}

	a.services = services

	log.Fatal("Server down: ", a.run())
}
