package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/api/routers"
	"github.com/mahdipeydai/taskmanager-go/api/validators"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/data/db"
	migration "github.com/mahdipeydai/taskmanager-go/data/db/migration"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/tests/mocks"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestSetup holds test environment and helpers
type TestSetup struct {
	Router      *gin.Engine
	Config      *config.Config
	Logger      logging.LoggerInterface
	DB          *gorm.DB
	RedisServer *miniredis.Miniredis
	T           *testing.T
}

// NewTestSetup initializes a test environment with a real database
func NewTestSetup(t *testing.T) *TestSetup {
	gin.SetMode(gin.TestMode)

	// Initialize mock redis
	redisServer, err := miniredis.Run()
	require.NoError(t, err)

	// Load config from environment or use defaults
	cfg := loadTestConfig(redisServer.Addr())

	// Initialize logger
	logger := &mocks.MockLogger{}

	val, ok := binding.Validator.Engine().(*validator.Validate)
	require.True(t, ok)

	err = val.RegisterValidation(
		"userPassword",
		validators.UserPasswordValidator(cfg),
		true,
	)
	require.NoError(t, err)

	// Get or initialize database connection
	err = db.InitDb(cfg, logger)
	require.NoError(t, err)

	database := db.GetDB()
	require.NotNil(t, database)

	migration.UpInit(cfg, logger)

	// Create Gin engine
	router := gin.New()
	router.Use(gin.Logger())

	return &TestSetup{
		Router:      router,
		Config:      cfg,
		Logger:      logger,
		DB:          database,
		RedisServer: redisServer,
		T:           t,
	}
}

// loadTestConfig creates a config for testing
func loadTestConfig(redisAddr string) *config.Config {
	host, port, _ := strings.Cut(redisAddr, ":")

	return &config.Config{
		Server: config.ServerConfig{
			InternalPort: 8080,
			ExternalPort: 8080,
			RateLimit:    100,
			AllowOrigins: "http://localhost:3000",
		},
		Postgres: config.PostgresConfig{
			Host:     "localhost",
			Port:     "5432",
			Username: "postgres",
			Password: "postgres",
			DbName:   "taskmanager_test",
		},
		Jwt: config.JwtConfig{
			Secret:                 "test-secret-key-for-jwt-testing-purposes-only",
			RefreshSecret:          "test-refresh-secret-key-for-jwt-testing",
			AccessTokenExpireTime:  3600,
			RefreshTokenExpireTime: 86400,
		},
		Redis: config.RedisConfig{
			Host: host,
			Port: port,
		},
		Logger: config.LoggerConfig{
			LoggerName: "zap",
		},
		Password: config.PasswordConfig{
			MaxLength:        32,
			MinLength:        8,
			IncludeChars:     true,
			IncludeDigits:    true,
			IncludeLowercase: true,
			IncludeUppercase: true,
		},
	}
}

// SetupRoutes registers all routes in the test engine
func (ts *TestSetup) SetupRoutes() {
	// Register the API routes
	apiGroup := ts.Router.Group("/api")
	v1Group := apiGroup.Group("/v1")

	// Health routes
	healthRouterGroup := v1Group.Group("/health")
	routers.HealthRouter(healthRouterGroup)

	// Users routes
	usersRouterGroup := v1Group.Group("/users")
	routers.UsersRouter(usersRouterGroup, ts.Config, ts.Logger)

	// Tasks routes
	tasksRouterGroup := v1Group.Group("/tasks")
	routers.TasksRouter(tasksRouterGroup, ts.Config, ts.Logger)
}

// Cleanup cleans up the test environment
func (ts *TestSetup) Cleanup() {
	if ts.RedisServer != nil {
		ts.RedisServer.Close()
	}
}

// SendRequest sends an HTTP request to the test router and returns response
func (ts *TestSetup) SendRequest(method, path string, body interface{}, token ...string) *httptest.ResponseRecorder {
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(ts.T, err)
		req, err = http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}

	require.NoError(ts.T, err)

	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	ts.Router.ServeHTTP(rec, req)

	return rec
}

// ParseResponse parses the response body into the given interface
func (ts *TestSetup) ParseResponse(rec *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(rec.Body.Bytes(), v)
	require.NoError(ts.T, err)
}

// HelperRegisterUser registers a test user and returns the tokens
func (ts *TestSetup) HelperRegisterUser(username, firstName, lastName, email, password string) dto.TokenDetail {
	registerReq := dto.RegisterUserByUsernameRequest{
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  password,
	}

	rec := ts.SendRequest("POST", "/api/v1/users/register", registerReq)
	if rec.Code != http.StatusCreated {
		ts.T.Logf("Registration failed with status %d: %s", rec.Code, rec.Body.String())
		ts.T.FailNow()
	}

	var response struct {
		Data dto.TokenDetail `json:"data"`
	}
	ts.ParseResponse(rec, &response)

	return response.Data
}

// HelperLoginUser logs in a user and returns the tokens
func (ts *TestSetup) HelperLoginUser(username, password string) dto.TokenDetail {
	loginReq := dto.LoginByUsernameRequest{
		Username: username,
		Password: password,
	}

	rec := ts.SendRequest("POST", "/api/v1/users/login", loginReq)
	require.Equal(ts.T, http.StatusOK, rec.Code, "login failed: "+rec.Body.String())

	var response struct {
		Data dto.TokenDetail `json:"data"`
	}
	ts.ParseResponse(rec, &response)

	return response.Data
}

// HelperCleanupUser deletes a test user from the database
func (ts *TestSetup) HelperCleanupUser(username string) {
	ts.DB.Where("username = ?", username).Delete(&map[string]interface{}{})
}

// HelperCleanupTask deletes a test task from the database
func (ts *TestSetup) HelperCleanupTask(taskID int) {
	ts.DB.Where("id = ?", taskID).Delete(&map[string]interface{}{})
}
