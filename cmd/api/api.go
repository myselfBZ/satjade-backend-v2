package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/myselfBZ/satjade-backend/internal/auth"
	"github.com/myselfBZ/satjade-backend/internal/filestore"
	auth_service "github.com/myselfBZ/satjade-backend/internal/services/auth"
	practices_service "github.com/myselfBZ/satjade-backend/internal/services/practices"
	questions_service "github.com/myselfBZ/satjade-backend/internal/services/questions"
	users_service "github.com/myselfBZ/satjade-backend/internal/services/users"
	"go.uber.org/zap"
)

type services struct {
	Auth      auth_service.AuthService
	Users     users_service.UserService
	Practices practices_service.PracticeService
	Questions questions_service.QuestionsService
}

type authConfig struct {
	secret string
	aud    string
	iss    string
	exp    time.Duration
}

type config struct {
	addr             string
	auth             authConfig
	frontEndUrl      string
	dbUrl            string
	secretKey        string
	refreshSecretKey string
	imgStorePath     string
	llmApiKey        string
}

type ErrEnvelope struct {
	Error string `json:"error"`
}

func (c *config) Load() {
	c.addr = os.Getenv("SERVER_PORT")
	if c.addr == "" {
		panic("SERVER_PORT was not set in the env")
	}

	c.frontEndUrl = os.Getenv("FRONTEND_URL")

	if c.frontEndUrl == "" {
		panic("FRONTEND_URL was not set in the env")
	}

	tokenExpHoursStr := os.Getenv("TOKEN_EXPR_HOURS")

	if tokenExpHoursStr == "" {
		panic("TOKEN_EXPR_HOURS was not set in the env")
	}

	tokenExpHours, err := strconv.Atoi(tokenExpHoursStr)
	if err != nil {
		panic("Please set a valid int for TOKEN_EXPR_HOURS")
	}

	c.auth.exp = time.Duration(tokenExpHours) * time.Hour

	c.auth.aud = os.Getenv("AUTH_AUD")

	if c.auth.aud == "" {
		panic("AUTH_AUD was not set in the env")
	}

	c.auth.iss = os.Getenv("AUTH_ISS")

	if c.auth.iss == "" {
		panic("AUTH_ISS was not set in the env")
	}

	c.dbUrl = os.Getenv("DB")

	if c.dbUrl == "" {
		panic("DB was not set in the env")
	}

	c.imgStorePath = os.Getenv("IMAGE_STORE")

	if c.imgStorePath == "" {
		panic("IMAGE_STORE was not set in the env")
	}

	c.llmApiKey = os.Getenv("LLM_APIKEY")

	if c.llmApiKey == "" {
		panic("LLM_APIKEY was not set in the env")
	}
}

type api struct {
	config      *config
	logger      *zap.SugaredLogger
	services    services
	validator   *validator.Validate
	auth        auth.Authenticator
	filestorage filestore.FileStorage
}

func (a *api) mount() *echo.Echo {
	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{a.config.frontEndUrl, "http://localhost:5174"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderAccessControlAllowOrigin,
		},
	}))

	v1 := e.Group("/v1")

	authR := v1.Group("/auth")

	authR.POST("/users", a.createAccountHandler)
	authR.POST("/login", a.loginHandler)

	usersR := v1.Group("/users", a.AuthMiddleware)

	usersR.GET("/self", a.getSelfHandler)
	usersR.GET("/:userId/questionattempts", a.getQuestionAttemptsByUserHandler)

	internalR := v1.Group("/internal", a.AuthMiddleware, a.CheckAdminMiddleware)
	internalR.GET("/users", a.getUsersHandler)
	internalR.POST("/questions", a.createQuestionHandler)
	internalR.POST("/modules/:moduleId/questions", a.createQuestionToModule)
	internalR.POST("/modules/:moduleId/questions/link", a.linkQuestionToModule)

	internalPracticeR := internalR.Group("/practices")
	internalPracticeR.POST("/", a.createPracticeHandler)
	internalPracticeR.POST("/:practiceId/publish", a.pulishPracticeHandler)
	internalPracticeR.GET("/:practiceId", a.getFullTest)

	practicesR := v1.Group("/practices", a.AuthMiddleware)
	practicesR.GET("/previews", a.getPracticePreviewsHandler)
	practicesR.GET("/published/:publishedId", a.getPublishedPracticeHandler)

	questions := v1.Group("/questions", a.AuthMiddleware)

	questions.GET("/:questionId", a.getQuestionHandler)

	questions.POST("/:questionId/attempts", a.createQuestionAttemptHandler)

	questions.GET("/ids", a.filterQuestionIdsHandler)
	questions.GET("/distribution", a.getQuestionDistributionHandler)

	v1.Static("/media", a.config.imgStorePath)
	return e
}

func (a *api) run() error {
	e := a.mount()
	return e.Start(a.config.addr)
}

func (a *api) getUUIDFromParam(name string, c echo.Context) (uuid.UUID, error) {
	id := c.Param(name)

	validID, err := uuid.Parse(id)

	if err != nil {
		return uuid.UUID{}, err
	}

	return validID, nil
}

func formatValidationError(err error) string {
	errors := ""

	valErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return "Internal validation error"
	}

	for _, f := range valErrors {
		var msg string

		switch f.Tag() {
		case "required":
			msg = fmt.Sprintf("%s is a required field. ", f.Field())
		case "email":
			msg = fmt.Sprintf("%s must be a valid email address. ", f.Field())
		case "gte":
			msg = fmt.Sprintf("%s must be at least %s. ", f.Field(), f.Param())
		case "lte":
			msg = fmt.Sprintf("%s cannot be greater than %s. ", f.Field(), f.Param())
		case "min":
			msg = fmt.Sprintf("%s must be at least %s characters. ", f.Field(), f.Param())
		case "max":
			msg = fmt.Sprintf("%s cannot exceed %s characters. ", f.Field(), f.Param())
		default:
			msg = fmt.Sprintf("%s is not valid input\n", f.Field())
		}

		errors += msg
	}

	return errors
}
