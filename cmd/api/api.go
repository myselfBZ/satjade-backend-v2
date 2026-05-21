package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/myselfBZ/satjade-backend/internal/auth"
	"github.com/myselfBZ/satjade-backend/internal/filestore"
	auth_service "github.com/myselfBZ/satjade-backend/internal/services/auth"
	friends_service "github.com/myselfBZ/satjade-backend/internal/services/friends"
	practiceattempt_service "github.com/myselfBZ/satjade-backend/internal/services/practice-attempt"
	practices_service "github.com/myselfBZ/satjade-backend/internal/services/practices"
	questions_service "github.com/myselfBZ/satjade-backend/internal/services/questions"
	users_service "github.com/myselfBZ/satjade-backend/internal/services/users"
	"github.com/myselfBZ/satjade-backend/internal/ws/challenge"
	"github.com/myselfBZ/satjade-backend/internal/ws/clients"
	"github.com/myselfBZ/satjade-backend/internal/ws/events"
	"go.uber.org/zap"
)

type services struct {
	Auth            auth_service.AuthService
	Users           users_service.UserService
	Practices       practices_service.PracticeService
	Questions       questions_service.QuestionsService
	PracticeAttempt practiceattempt_service.PracticeAttemptService
	Friends         friends_service.FriendsService
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
	logFile			 string
}

func (c *config) Load() {
	c.logFile = os.Getenv("LOG_FILE")

	if c.logFile == "" {
		panic("LOG_FILE was not set in the env")
	}

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

	wsReadCh       chan events.ClientSentEvent
	wsClients      clients.Manager
	wsClientExitCh chan string
	challenges     challenge.Manager

	eventCh chan eventWrapper
	duels   *duelMap
}

func (a *api) mount() *echo.Echo {
	e := echo.New()

	// considering it's importance....
	// e.Use(middleware.RequestLogger())
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
	usersR.GET("/me/friends", a.getFriendsHandler)
	usersR.POST("/:user_id/friends", a.sendFriendshipRequestHandler)

	usersR.GET("/:userId/questionattempts", a.getQuestionAttemptsByUserHandler)
	usersR.GET("/search", a.searchUserHandler)

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
	practicesR.GET("/published", a.getPublishedPracticesPreviewsHandler)

	attemptsR := v1.Group("/attempts", a.AuthMiddleware)
	attemptsR.POST("/", a.createPracticeAttemptHandler)
	attemptsR.GET("/", a.getPracticeAttemptPreviewsHandler)
	attemptsR.GET("/:attemptId", a.getPracticeAttemptByIdHandler)

	questions := v1.Group("/questions", a.AuthMiddleware)

	questions.GET("/:questionId", a.getQuestionHandler)

	questions.POST("/:questionId/attempts", a.createQuestionAttemptHandler)

	questions.GET("/ids", a.filterQuestionIdsHandler)
	questions.GET("/distribution", a.getQuestionDistributionHandler)

	friendshipR := v1.Group("/friendship", a.AuthMiddleware)
	friendshipR.GET("/requests", a.getFriendshipRequestsHandler)
	friendshipR.DELETE("/requests/:request_id", a.rejectFriendshipRequestHandler)
	friendshipR.POST("/requests/:request_id", a.acceptFriendshipRequestHandler)
	friendshipR.DELETE("/:id", a.deleteFriendHandler)

	v1.GET("/ws/duel", a.handleWSConn)

	v1.Static("/media", a.config.imgStorePath)
	return e
}

func (a *api) run() error {
	e := a.mount()
	go a.handleEventLoop()
	go a.onDissconnect()
	// TODO do something with the expired challenges, but do i really have to?
	go func() {
		for req := range a.challenges.ExpiredCh() {
			_ = req
			a.logger.Logw(zap.DebugLevel, "Challenge expired")
		}
	}()

	// TODO make it configurable
	s := &http.Server{
		Addr:              a.config.addr,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return e.StartServer(s)
}
