package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTasksEndpoint_Create tests task creation
func TestTasksEndpoint_Create(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	// Register and login a user to get token
	suffix := time.Now().UnixNano() % 1_000_000
	username := fmt.Sprintf("taskuser%d", suffix)
	email := fmt.Sprintf("taskuser%d@example.com", suffix)
	testUser := dto.RegisterUserByUsernameRequest{
		Username:  username,
		FirstName: "Task",
		LastName:  "Tester",
		Email:     email,
		Password:  "SecurePass123!",
	}
	tokens := setup.HelperRegisterUser(testUser.Username, testUser.FirstName, testUser.LastName, testUser.Email, testUser.Password)
	defer setup.HelperCleanupUser(testUser.Username)

	tests := []struct {
		name           string
		request        dto.CreateTaskRequest
		token          string
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "successful task creation",
			request: dto.CreateTaskRequest{
				Title:       "Write integration tests",
				Description: "Create comprehensive integration tests for all API endpoints",
				Status:      models.TaskStatusPending,
			},
			token:          tokens.AccessToken,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.NotZero(t, data["id"])
				assert.Equal(t, "Write integration tests", data["title"])
			},
		},
		{
			name: "task creation without authentication",
			request: dto.CreateTaskRequest{
				Title:       "Unauthorized task",
				Description: "This should fail",
				Status:      models.TaskStatusPending,
			},
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "task creation with invalid token",
			request: dto.CreateTaskRequest{
				Title:       "Invalid token task",
				Description: "This should fail",
				Status:      models.TaskStatusPending,
			},
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "task creation without title",
			request: dto.CreateTaskRequest{
				Title:       "",
				Description: "This should fail",
				Status:      models.TaskStatusPending,
			},
			token:          tokens.AccessToken,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name: "task creation without description",
			request: dto.CreateTaskRequest{
				Title:  "No description task",
				Status: models.TaskStatusPending,
			},
			token:          tokens.AccessToken,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("POST", "/api/v1/tasks", tt.request, tt.token)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}

// TestTasksEndpoint_GetByID tests retrieving a specific task
func TestTasksEndpoint_GetByID(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	// Register and login a user
	suffix := time.Now().UnixNano() % 1_000_000
	username := fmt.Sprintf("taskuser%d", suffix)
	email := fmt.Sprintf("taskuser%d@example.com", suffix)
	testUser := dto.RegisterUserByUsernameRequest{
		Username:  username,
		FirstName: "Task",
		LastName:  "Tester",
		Email:     email,
		Password:  "SecurePass123!",
	}
	tokens := setup.HelperRegisterUser(testUser.Username, testUser.FirstName, testUser.LastName, testUser.Email, testUser.Password)

	// Create a task
	createReq := dto.CreateTaskRequest{
		Title:       "Test task for retrieval",
		Description: "Task to be retrieved",
		Status:      models.TaskStatusPending,
	}
	createRec := setup.SendRequest("POST", "/api/v1/tasks", createReq, tokens.AccessToken)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResponse struct {
		Data dto.TaskResponse `json:"data"`
	}
	setup.ParseResponse(createRec, &createResponse)
	taskID := createResponse.Data.ID

	tests := []struct {
		name           string
		taskID         int
		token          string
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "successful task retrieval",
			taskID:         taskID,
			token:          tokens.AccessToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.Equal(t, float64(taskID), data["id"])
			},
		},
		{
			name:           "task retrieval without authentication",
			taskID:         taskID,
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name:           "non-existent task",
			taskID:         99999,
			token:          tokens.AccessToken,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("GET", fmt.Sprintf("/api/v1/tasks/%d", tt.taskID), nil, tt.token)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}

// TestTasksEndpoint_GetByFilter tests retrieving tasks with filters
func TestTasksEndpoint_GetByFilter(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	// Register and login a user
	suffix := time.Now().UnixNano() % 1_000_000
	username := fmt.Sprintf("taskuser%d", suffix)
	email := fmt.Sprintf("taskuser%d@example.com", suffix)
	testUser := dto.RegisterUserByUsernameRequest{
		Username:  username,
		FirstName: "Task",
		LastName:  "Tester",
		Email:     email,
		Password:  "SecurePass123!",
	}
	tokens := setup.HelperRegisterUser(testUser.Username, testUser.FirstName, testUser.LastName, testUser.Email, testUser.Password)

	// Create multiple tasks with different statuses
	tasks := []dto.CreateTaskRequest{
		{
			Title:       "Pending task 1",
			Description: "First pending task",
			Status:      models.TaskStatusPending,
		},
		{
			Title:       "In progress task",
			Description: "Task in progress",
			Status:      models.TaskStatusInProgress,
		},
		{
			Title:       "Completed task",
			Description: "Already completed",
			Status:      models.TaskStatusCompleted,
		},
	}

	for _, taskReq := range tasks {
		setup.SendRequest("POST", "/api/v1/tasks", taskReq, tokens.AccessToken)
	}

	tests := []struct {
		name           string
		query          string
		token          string
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "list all tasks",
			query:          "?page=1&pageSize=10",
			token:          tokens.AccessToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				items := data["items"].([]interface{})
				assert.GreaterOrEqual(t, len(items), 3)
			},
		},
		{
			name:           "list without authentication",
			query:          "?page=1&pageSize=10",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name:           "filter tasks by status",
			query:          "?status=pending&page=1&pageSize=10",
			token:          tokens.AccessToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("GET", "/api/v1/tasks"+tt.query, nil, tt.token)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}

// TestTasksEndpoint_Update tests task updates
func TestTasksEndpoint_Update(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	// Register and login a user
	suffix := time.Now().UnixNano() % 1_000_000
	username := fmt.Sprintf("taskuser%d", suffix)
	email := fmt.Sprintf("taskuser%d@example.com", suffix)
	testUser := dto.RegisterUserByUsernameRequest{
		Username:  username,
		FirstName: "Task",
		LastName:  "Tester",
		Email:     email,
		Password:  "SecurePass123!",
	}
	tokens := setup.HelperRegisterUser(testUser.Username, testUser.FirstName, testUser.LastName, testUser.Email, testUser.Password)

	// Create a task
	createReq := dto.CreateTaskRequest{
		Title:       "Task to be updated",
		Description: "Original description",
		Status:      models.TaskStatusPending,
	}
	createRec := setup.SendRequest("POST", "/api/v1/tasks", createReq, tokens.AccessToken)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResponse struct {
		Data dto.TaskResponse `json:"data"`
	}
	setup.ParseResponse(createRec, &createResponse)
	taskID := createResponse.Data.ID

	tests := []struct {
		name           string
		taskID         int
		request        dto.UpdateTaskRequest
		token          string
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:   "successful task update",
			taskID: taskID,
			request: dto.UpdateTaskRequest{
				Title:       strPtr("Updated title"),
				Description: strPtr("Updated description"),
				Status:      statusPtr(models.TaskStatusInProgress),
			},
			token:          tokens.AccessToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
				data := response["data"].(map[string]interface{})
				assert.Equal(t, "Updated title", data["title"])
			},
		},
		{
			name:           "update without authentication",
			taskID:         taskID,
			request:        dto.UpdateTaskRequest{},
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name:   "partial task update",
			taskID: taskID,
			request: dto.UpdateTaskRequest{
				Status: statusPtr(models.TaskStatusCompleted),
			},
			token:          tokens.AccessToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
			},
		},
		{
			name:           "update non-existent task",
			taskID:         99999,
			request:        dto.UpdateTaskRequest{},
			token:          tokens.AccessToken,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("PUT", fmt.Sprintf("/api/v1/tasks/%d", tt.taskID), tt.request, tt.token)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}

// TestTasksEndpoint_Delete tests task deletion
func TestTasksEndpoint_Delete(t *testing.T) {
	setup := NewTestSetup(t)
	defer setup.Cleanup()
	setup.SetupRoutes()

	// Register and login a user
	suffix := time.Now().UnixNano() % 1_000_000
	username := fmt.Sprintf("taskuser%d", suffix)
	email := fmt.Sprintf("taskuser%d@example.com", suffix)
	testUser := dto.RegisterUserByUsernameRequest{
		Username:  username,
		FirstName: "Task",
		LastName:  "Tester",
		Email:     email,
		Password:  "SecurePass123!",
	}
	tokens := setup.HelperRegisterUser(testUser.Username, testUser.FirstName, testUser.LastName, testUser.Email, testUser.Password)

	// Create a task to delete
	createReq := dto.CreateTaskRequest{
		Title:       "Task to be deleted",
		Description: "This task will be deleted",
		Status:      models.TaskStatusPending,
	}
	createRec := setup.SendRequest("POST", "/api/v1/tasks", createReq, tokens.AccessToken)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var createResponse struct {
		Data dto.TaskResponse `json:"data"`
	}
	setup.ParseResponse(createRec, &createResponse)
	taskID := createResponse.Data.ID

	tests := []struct {
		name           string
		taskID         int
		token          string
		expectedStatus int
		checkResponse  func(t *testing.T, response map[string]interface{})
	}{
		{
			name:           "successful task deletion",
			taskID:         taskID,
			token:          tokens.AccessToken,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
			},
		},
		{
			name:           "delete without authentication",
			taskID:         taskID,
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
		{
			name:           "delete non-existent task",
			taskID:         99999,
			token:          tokens.AccessToken,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.False(t, response["success"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := setup.SendRequest("DELETE", fmt.Sprintf("/api/v1/tasks/%d", tt.taskID), nil, tt.token)

			assert.Equal(t, tt.expectedStatus, rec.Code, "response body: "+rec.Body.String())

			var response map[string]interface{}
			setup.ParseResponse(rec, &response)

			tt.checkResponse(t, response)
		})
	}
}

// Helper functions for pointer types
func strPtr(s string) *string {
	return &s
}

func statusPtr(s models.TaskStatus) *models.TaskStatus {
	return &s
}
