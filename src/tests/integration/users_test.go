package integration

import (
	"net/http"
	"testing"

	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/stretchr/testify/assert"
)

// TestUsersEndpoint_Register tests user registration
func TestUsersEndpoint_Register(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	tests := []struct {
		name           string
		request        dto.RegisterUserByUsernameRequest
		expectedStatus int
		checkResponse  func(t *testing.T, rec interface{}, shouldHaveToken bool)
	}{
		{
			name: "successful registration",
			request: dto.RegisterUserByUsernameRequest{
				Username:  "testuser1",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "SecurePass123!",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec interface{}, shouldHaveToken bool) {
				response := rec.(map[string]interface{})
				if shouldHaveToken {
					data := response["data"].(map[string]interface{})
					assert.NotEmpty(t, data["accessToken"])
					assert.NotEmpty(t, data["refreshToken"])
				}
			},
		},
		{
			name: "missing required fields",
			request: dto.RegisterUserByUsernameRequest{
				Username: "testuser2",
				// Missing FirstName
				LastName: "Doe",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec interface{}, shouldHaveToken bool) {
				response := rec.(map[string]interface{})
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "invalid email format",
			request: dto.RegisterUserByUsernameRequest{
				Username:  "testuser3",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "invalid-email",
				Password:  "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec interface{}, shouldHaveToken bool) {
				response := rec.(map[string]interface{})
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "weak password",
			request: dto.RegisterUserByUsernameRequest{
				Username:  "testuser4",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john4@example.com",
				Password:  "weak",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec interface{}, shouldHaveToken bool) {
				response := rec.(map[string]interface{})
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("POST", "/api/v1/users/register", tt.request)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response, tt.expectedStatus == http.StatusCreated)

			// Cleanup if registration succeeded
			if tt.expectedStatus == http.StatusCreated {
				setup.HelperCleanupUser(tt.request.Username)
			}
		})
	}
}

// TestUsersEndpoint_Login tests user login
func TestUsersEndpoint_Login(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	// First register a user
	testUser := dto.RegisterUserByUsernameRequest{
		Username:  "logintest",
		FirstName: "Test",
		LastName:  "User",
		Email:     "logintest@example.com",
		Password:  "SecurePass123!",
	}
	setup.HelperRegisterUser(testUser.Username, testUser.FirstName, testUser.LastName, testUser.Email, testUser.Password)
	defer setup.HelperCleanupUser(testUser.Username)

	tests := []struct {
		name           string
		request        dto.LoginByUsernameRequest
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "successful login",
			request: dto.LoginByUsernameRequest{
				Username: "logintest",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["accessToken"])
				assert.NotEmpty(t, data["refreshToken"])
				assert.NotZero(t, data["accessTokenExpireTime"])
				assert.NotZero(t, data["refreshTokenExpireTime"])
			},
		},
		{
			name: "invalid password",
			request: dto.LoginByUsernameRequest{
				Username: "logintest",
				Password: "WrongPassword123!",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "non-existent user",
			request: dto.LoginByUsernameRequest{
				Username: "nonexistent",
				Password: "SomePassword123!",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "missing password",
			request: dto.LoginByUsernameRequest{
				Username: "logintest",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("POST", "/api/v1/users/login", tt.request)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}

// TestUsersEndpoint_RefreshToken tests token refresh
func TestUsersEndpoint_RefreshToken(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	// First register and login a user
	testUser := dto.RegisterUserByUsernameRequest{
		Username:  "refreshtest",
		FirstName: "Test",
		LastName:  "User",
		Email:     "refreshtest@example.com",
		Password:  "SecurePass123!",
	}
	tokens := setup.HelperRegisterUser(testUser.Username, testUser.FirstName, testUser.LastName, testUser.Email, testUser.Password)
	defer setup.HelperCleanupUser(testUser.Username)

	oldRefreshToken := tokens.RefreshToken

	tests := []struct {
		name           string
		request        dto.RefreshTokenRequest
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "successful token refresh",
			request: dto.RefreshTokenRequest{
				RefreshToken: oldRefreshToken,
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["accessToken"])
				assert.NotEmpty(t, data["refreshToken"])
			},
		},
		{
			name: "invalid refresh token",
			request: dto.RefreshTokenRequest{
				RefreshToken: "invalid.token.here",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "empty refresh token",
			request: dto.RefreshTokenRequest{
				RefreshToken: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("POST", "/api/v1/users/refresh-token", tt.request)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}
