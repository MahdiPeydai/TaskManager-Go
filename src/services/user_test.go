package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/common"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/data/models"
	"github.com/mahdipeydai/taskmanager-go/pkg/service_errors"
	"github.com/mahdipeydai/taskmanager-go/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testUsersService(t *testing.T) (*UsersService, sqlmock.Sqlmock) {
	t.Helper()

	cfg := &config.Config{}

	cfg.Jwt.Secret = "test-secret"
	cfg.Jwt.AccessTokenExpireTime = 3600
	cfg.Jwt.RefreshTokenExpireTime = 86400

	mockDB, err := mocks.NewMockDatabase()
	require.NoError(t, err)

	logger := &mocks.MockLogger{}

	service := GetUsersService(
		mockDB.DB,
		cfg,
		logger,
	)

	return service, mockDB.Mock
}

func createRefreshTokenForTest(
	t *testing.T,
	service *TokenService,
	userID int,
) string {
	t.Helper()
	ctx := context.Background()

	result, err := service.GenerateToken(
		ctx,
		Token{
			UserId:   userID,
			Username: "john",
		},
	)

	require.NoError(t, err)
	require.NotEmpty(t, result.RefreshToken)

	return result.RefreshToken
}

func testRegisterRequest() *dto.RegisterUserByUsernameRequest {
	return &dto.RegisterUserByUsernameRequest{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "john",
		Email:     "john@example.com",
		Password:  "password123",
	}
}

func TestUsersService_CreateUserToken(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	firstName := "John"
	lastName := "Doe"
	email := "john@example.com"
	mobile := "09123456789"

	user := &models.User{
		BaseModel: models.BaseModel{
			Id: 123,
		},
		Username:     "john",
		Firstname:    &firstName,
		Lastname:     &lastName,
		Email:        &email,
		MobileNumber: &mobile,
		Roles: []models.UserRole{
			{
				Role: &models.Role{
					Name: "user",
				},
			},
			{
				Role: &models.Role{
					Name: "admin",
				},
			},
		},
	}

	result, err := service.CreateUserToken(ctx, user)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RefreshUserToken(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	refreshToken := createRefreshTokenForTest(t, service.tokenService, 123)

	mock.ExpectQuery(
		`SELECT \* FROM "users" WHERE id = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`,
	).
		WithArgs(float64(123), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"username",
				"first_name",
				"last_name",
				"email",
				"is_active",
				"mobile_verified",
			}).AddRow(
				123,
				"john",
				"John",
				"Doe",
				"john@example.com",
				true,
				false,
			),
		)

	// UserRole preload.
	mock.ExpectQuery(
		`SELECT \* FROM "user_roles" WHERE "user_roles"\."user_id" = \$1 AND "user_roles"\."deleted_at" IS NULL`,
	).
		WithArgs(123).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"role_id",
			}).AddRow(
				1,
				123,
				1,
			),
		)

	// Role preload.
	mock.ExpectQuery(
		`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1 AND "roles"\."deleted_at" IS NULL`,
	).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"name",
			}).AddRow(
				1,
				"user",
			),
		)

	result, err := service.RefreshUserToken(
		ctx,
		&dto.RefreshTokenRequest{
			RefreshToken: refreshToken,
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RefreshUserToken_InvalidToken(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	result, err := service.RefreshUserToken(
		ctx,
		&dto.RefreshTokenRequest{
			RefreshToken: "invalid-token",
		},
	)

	assert.Nil(t, result)
	assert.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RefreshUserToken_AccessToken(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	token, err := service.tokenService.GenerateToken(
		ctx,
		Token{
			UserId:   123,
			Username: "john",
		},
	)

	require.NoError(t, err)

	accessToken := strings.TrimPrefix(
		token.AccessToken,
		constants.AuthorizationHeaderPrefix+" ",
	)

	result, err := service.RefreshUserToken(
		ctx,
		&dto.RefreshTokenRequest{
			RefreshToken: accessToken,
		},
	)

	assert.Nil(t, result)
	require.Error(t, err)

	var serviceErr service_errors.ServiceError
	require.ErrorAs(t, err, &serviceErr)
	assert.Equal(
		t,
		service_errors.InvalidToken,
		serviceErr.EndUserMessage,
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RefreshUserToken_UserNotFound(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	refreshToken := createRefreshTokenForTest(
		t,
		service.tokenService,
		123,
	)

	mock.ExpectQuery(
		`SELECT \* FROM "users" WHERE id = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`,
	).
		WithArgs(float64(123), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := service.RefreshUserToken(
		ctx,
		&dto.RefreshTokenRequest{
			RefreshToken: refreshToken,
		},
	)

	assert.Nil(t, result)
	require.Error(t, err)

	var serviceErr service_errors.ServiceError
	require.ErrorAs(t, err, &serviceErr)

	assert.Equal(
		t,
		service_errors.RecordNotFound,
		serviceErr.EndUserMessage,
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_LoginByUsername(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	password := "password123"

	hashedPassword, err := common.HashPassword(password)
	require.NoError(t, err)

	mock.ExpectQuery(
		`SELECT \* FROM "users" WHERE username = \$1 AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2`,
	).
		WithArgs("john", 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"username",
				"password",
			}).AddRow(
				123,
				"john",
				hashedPassword,
			),
		)

	// Then expect UserRole preload.
	mock.ExpectQuery(
		`SELECT \* FROM "user_roles" WHERE "user_roles"\."user_id" = \$1`,
	).
		WithArgs(123).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"user_id",
				"role_id",
			}).AddRow(1, 123, 1),
		)

	// Then Role preload.
	mock.ExpectQuery(
		`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`,
	).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"name",
			}).AddRow(1, "user"),
		)

	result, err := service.LoginByUsername(
		ctx,
		&dto.LoginByUsernameRequest{
			Username: "john",
			Password: password,
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_Success(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	// Username does not exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email does not exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Default role exists.
	mock.ExpectQuery(
		`SELECT "id" FROM "roles" WHERE name = \$1 AND "roles"\."deleted_at" IS NULL ORDER BY "roles"\."id" LIMIT \$2`,
	).
		WithArgs(constants.DefaultRoleName, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

	// Transaction begins.
	mock.ExpectBegin()

	// User is created.
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(10),
		)

	// UserRole is created.
	mock.ExpectQuery(`INSERT INTO "user_roles"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(20),
		)

	// Transaction commits.
	mock.ExpectCommit()

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_UsernameExists(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(true),
		)

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)

	var serviceErr service_errors.ServiceError
	require.ErrorAs(t, err, &serviceErr)

	assert.Equal(
		t,
		service_errors.UsernameExists,
		serviceErr.EndUserMessage,
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_EmailExists(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	// Username does not exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email already exists.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(true),
		)

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)

	var serviceErr service_errors.ServiceError
	require.ErrorAs(t, err, &serviceErr)

	assert.Equal(
		t,
		service_errors.EmailExists,
		serviceErr.EndUserMessage,
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_UsernameQueryDatabaseError(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	dbError := errors.New("database connection error")

	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnError(dbError)

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_EmailQueryDatabaseError(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	dbError := errors.New("database connection error")

	// Username doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email query fails.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnError(dbError)

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_DefaultRoleNotFound(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	// Username doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Default role doesn't exist.
	mock.ExpectQuery(
		`SELECT "id" FROM "roles" WHERE name = \$1 AND "roles"\."deleted_at" IS NULL ORDER BY "roles"\."id" LIMIT \$2`,
	).
		WithArgs(constants.DefaultRoleName, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_TransactionBeginFailure(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	// Username doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Default role exists.
	mock.ExpectQuery(
		`SELECT "id" FROM "roles" WHERE name = \$1 AND "roles"\."deleted_at" IS NULL ORDER BY "roles"\."id" LIMIT \$2`,
	).
		WithArgs(constants.DefaultRoleName, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

	dbError := errors.New("failed to begin transaction")

	mock.ExpectBegin().WillReturnError(dbError)

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_CreateUserFailure(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	// Username doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Default role exists.
	mock.ExpectQuery(
		`SELECT "id" FROM "roles" WHERE name = \$1 AND "roles"\."deleted_at" IS NULL ORDER BY "roles"\."id" LIMIT \$2`,
	).
		WithArgs(constants.DefaultRoleName, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

	mock.ExpectBegin()

	dbError := errors.New("failed to create user")

	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(dbError)

	mock.ExpectRollback()

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_CreateUserRoleFailure(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	// Username doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Default role exists.
	mock.ExpectQuery(
		`SELECT "id" FROM "roles" WHERE name = \$1 AND "roles"\."deleted_at" IS NULL ORDER BY "roles"\."id" LIMIT \$2`,
	).
		WithArgs(constants.DefaultRoleName, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

	mock.ExpectBegin()

	// User creation succeeds.
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(10),
		)

	// UserRole creation fails.
	dbError := errors.New("failed to create user role")

	mock.ExpectQuery(`INSERT INTO "user_roles"`).
		WillReturnError(dbError)

	mock.ExpectRollback()

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_RegisterByUsername_CommitFailure(t *testing.T) {
	service, mock := testUsersService(t)
	ctx := context.Background()

	// Username doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE username = \$1`,
	).
		WithArgs("john").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Email doesn't exist.
	mock.ExpectQuery(
		`SELECT count\(\*\) > 0 FROM "users" WHERE email = \$1`,
	).
		WithArgs("john@example.com").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(false),
		)

	// Default role exists.
	mock.ExpectQuery(
		`SELECT "id" FROM "roles" WHERE name = \$1 AND "roles"\."deleted_at" IS NULL ORDER BY "roles"\."id" LIMIT \$2`,
	).
		WithArgs(constants.DefaultRoleName, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(1),
		)

	mock.ExpectBegin()

	// User creation succeeds.
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(10),
		)

	// UserRole creation succeeds.
	mock.ExpectQuery(`INSERT INTO "user_roles"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(20),
		)

	// Commit fails.
	dbError := errors.New("commit failed")
	mock.ExpectCommit().WillReturnError(dbError)

	err := service.RegisterByUsername(ctx, testRegisterRequest())

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)

	require.NoError(t, mock.ExpectationsWereMet())
}
