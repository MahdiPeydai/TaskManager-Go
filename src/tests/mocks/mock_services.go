package mocks

import (
	"context"

	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/stretchr/testify/mock"
)

// MockUsersService is a mock implementation of UsersService
type MockUsersService struct {
	mock.Mock
}

func (m *MockUsersService) RegisterByUsername(ctx context.Context, req *dto.RegisterUserByUsernameRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUsersService) LoginByUsername(ctx context.Context, req *dto.LoginByUsernameRequest) (*dto.TokenDetail, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenDetail), args.Error(1)
}

func (m *MockUsersService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenDetail, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenDetail), args.Error(1)
}

// MockTaskService is a mock implementation of TaskService
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) Create(ctx context.Context, req *dto.CreateTaskRequest, userID int) (*dto.TaskResponse, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TaskResponse), args.Error(1)
}

func (m *MockTaskService) GetByID(ctx context.Context, id int, userID int) (*dto.TaskResponse, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TaskResponse), args.Error(1)
}

func (m *MockTaskService) GetByFilter(ctx context.Context, req *dto.TaskListRequest, userID int) (*dto.TaskListResponse, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TaskListResponse), args.Error(1)
}

func (m *MockTaskService) Update(ctx context.Context, id int, req *dto.UpdateTaskRequest, userID int) (*dto.TaskResponse, error) {
	args := m.Called(ctx, id, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TaskResponse), args.Error(1)
}

func (m *MockTaskService) Delete(ctx context.Context, id int, userID int) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// MockTokenService is a mock implementation of TokenService
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GetClaims(ctx context.Context, token string) (map[string]interface{}, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// Helper function to create a token with claims
func CreateMockToken(userID int, username string, role string) string {
	// This is a placeholder - in real tests, use JWT library to create valid tokens
	return "mock.token.here"
}

// Helper function to create mock token claims
func CreateMockClaims(userID int, username string, role string) map[string]interface{} {
	return map[string]interface{}{
		"user_id":  float64(userID),
		"username": username,
		"role":     role,
		"exp":      float64(9999999999),
	}
}
