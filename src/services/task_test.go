package services

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/data/models"
	"github.com/mahdipeydai/taskmanager-go/pkg/service_errors"
	"github.com/mahdipeydai/taskmanager-go/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func testTaskService(t *testing.T) (*TaskService, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()

	cfg := &config.Config{}

	mockDB, err := mocks.NewMockDatabase()
	require.NoError(t, err)

	logger := &mocks.MockLogger{}

	redisServer := miniredis.RunT(t)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	service := GetTaskService(
		mockDB.DB,
		redisClient,
		cfg,
		logger,
	)

	return service, mockDB.Mock, redisServer
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	service, mock, redisServer := testTaskService(t)

	ctx := context.Background()

	req := &dto.CreateTaskRequest{
		Title:       "Test task",
		Description: "Test description",
		Status:      models.TaskStatusPending,
	}

	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO "tasks"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(10),
		)

	mock.ExpectCommit()

	// GetByID: cache miss.
	// miniredis starts empty, so cache.Get returns redis.Nil.

	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
		),
	).
		WithArgs(10, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
				"deleted_at",
			}).
				AddRow(
					10,
					"Test task",
					"Test description",
					"pending",
					nil,
					nil,
				),
		)

	result, err := service.CreateTask(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 10, result.ID)
	assert.Equal(t, "Test task", result.Title)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, models.TaskStatusPending, result.Status)

	require.NoError(t, mock.ExpectationsWereMet())
	assert.True(t, redisServer.Exists("task:10"))
}

func TestTaskService_CreateTask_CreateFailure(t *testing.T) {
	service, mock, _ := testTaskService(t)

	ctx := context.Background()

	dbError := errors.New("create failed")

	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO "tasks"`).
		WillReturnError(dbError)

	mock.ExpectRollback()

	result, err := service.CreateTask(
		ctx,
		&dto.CreateTaskRequest{
			Title: "Test task",
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_CreateTask_CommitFailure(t *testing.T) {
	service, mock, _ := testTaskService(t)

	ctx := context.Background()

	dbError := errors.New("commit failed")

	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO "tasks"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(10),
		)

	mock.ExpectCommit().
		WillReturnError(dbError)

	result, err := service.CreateTask(
		ctx,
		&dto.CreateTaskRequest{
			Title: "Test task",
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_CreateTask_BeginFailure(t *testing.T) {
	service, mock, _ := testTaskService(t)

	dbError := errors.New("begin failed")

	mock.ExpectBegin().
		WillReturnError(dbError)

	result, err := service.CreateTask(
		context.Background(),
		&dto.CreateTaskRequest{
			Title: "Test task",
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_GetByID_CacheMiss(t *testing.T) {
	service, mock, redisServer := testTaskService(t)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
		),
	).
		WithArgs(10, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
				"deleted_at",
			}).
				AddRow(
					10,
					"Test task",
					"Test description",
					"pending",
					nil,
					nil,
				),
		)

	result, err := service.GetByID(context.Background(), 10)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 10, result.ID)
	assert.Equal(t, "Test task", result.Title)

	require.NoError(t, mock.ExpectationsWereMet())

	// GetByID should cache the database result.
	assert.True(t, redisServer.Exists("task:10"))
}

func TestTaskService_GetByID_NotFound(t *testing.T) {
	service, mock, _ := testTaskService(t)

	query := regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := service.GetByID(context.Background(), 10)

	require.Error(t, err)
	assert.Nil(t, result)

	var serviceErr service_errors.ServiceError
	require.ErrorAs(t, err, &serviceErr)
	assert.Equal(
		t,
		service_errors.RecordNotFound,
		serviceErr.EndUserMessage,
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_GetByID_DatabaseError(t *testing.T) {
	service, mock, _ := testTaskService(t)

	dbError := errors.New("database unavailable")

	query := regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnError(dbError)

	result, err := service.GetByID(context.Background(), 10)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_Update_Success(t *testing.T) {
	service, mock, _ := testTaskService(t)

	ctx := context.WithValue(
		context.Background(),
		constants.UserIdKey,
		float64(99),
	)

	title := "Updated task"
	description := "Updated description"
	status := models.TaskStatusCompleted
	assigneeID := 20

	mock.ExpectBegin()

	query := regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
			}).AddRow(
				10,
				"Old task",
				"Old description",
				models.TaskStatusPending,
				10,
			),
		)

	mock.ExpectExec(
		`UPDATE "tasks"`,
	).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectCommit()

	// GetByID after cache invalidation.
	query = regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
			}).AddRow(
				10,
				title,
				description,
				status,
				assigneeID,
			),
		)

	result, err := service.Update(
		ctx,
		10,
		&dto.UpdateTaskRequest{
			Title:       &title,
			Description: &description,
			Status:      &status,
			AssigneeID:  &assigneeID,
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 10, result.ID)
	assert.Equal(t, title, result.Title)
	assert.Equal(t, description, result.Description)
	assert.Equal(t, status, result.Status)
	assert.Equal(t, assigneeID, *result.AssigneeID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_Update_NotFound(t *testing.T) {
	service, mock, _ := testTaskService(t)

	ctx := context.WithValue(
		context.Background(),
		constants.UserIdKey,
		float64(99),
	)

	mock.ExpectBegin()

	query := regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectRollback()

	result, err := service.Update(
		ctx,
		10,
		&dto.UpdateTaskRequest{},
	)

	require.Error(t, err)
	assert.Nil(t, result)

	var serviceErr service_errors.ServiceError
	require.ErrorAs(t, err, &serviceErr)

	assert.Equal(
		t,
		service_errors.RecordNotFound,
		serviceErr.EndUserMessage,
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_Update_SelectError(t *testing.T) {
	service, mock, _ := testTaskService(t)

	ctx := context.WithValue(
		context.Background(),
		constants.UserIdKey,
		float64(99),
	)

	dbError := errors.New("select failed")

	mock.ExpectBegin()

	query := regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnError(dbError)

	mock.ExpectRollback()

	result, err := service.Update(
		ctx,
		10,
		&dto.UpdateTaskRequest{},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_Update_UpdateError(t *testing.T) {
	service, mock, _ := testTaskService(t)

	ctx := context.WithValue(
		context.Background(),
		constants.UserIdKey,
		float64(99),
	)

	title := "Updated"
	dbError := errors.New("update failed")

	mock.ExpectBegin()

	query := regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE (id = $1 AND deleted_by is null) AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
			}).AddRow(
				10,
				"Old",
			),
		)

	mock.ExpectExec(`UPDATE "tasks"`).
		WillReturnError(dbError)

	mock.ExpectRollback()

	result, err := service.Update(
		ctx,
		10,
		&dto.UpdateTaskRequest{
			Title: &title,
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_Delete_NotFound(t *testing.T) {
	service, mock, _ := testTaskService(t)

	query := regexp.QuoteMeta(
		`SELECT * FROM "tasks" WHERE id = $1 AND "tasks"."deleted_at" IS NULL ORDER BY "tasks"."id" LIMIT $2`,
	)

	mock.ExpectQuery(query).
		WithArgs(10, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	err := service.Delete(
		context.WithValue(
			context.Background(),
			constants.UserIdKey,
			float64(99),
		),
		10,
	)

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

func TestTaskService_Delete_SelectError(t *testing.T) {
	service, mock, _ := testTaskService(t)

	dbError := errors.New("database failed")

	mock.ExpectQuery(
		`SELECT \* FROM "tasks" WHERE id = \$1`,
	).
		WithArgs(10, 1).
		WillReturnError(dbError)

	err := service.Delete(
		context.WithValue(
			context.Background(),
			constants.UserIdKey,
			float64(99),
		),
		10,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_Delete_Success(t *testing.T) {
	service, mock, redisServer := testTaskService(t)

	ctx := context.WithValue(
		context.Background(),
		constants.UserIdKey,
		float64(99),
	)

	// Initial existence check.
	mock.ExpectQuery(
		`SELECT \* FROM "tasks" WHERE id = \$1`,
	).
		WithArgs(10, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
			}).AddRow(
				10,
				"Task",
			),
		)

	mock.ExpectBegin()

	mock.ExpectExec(`UPDATE "tasks"`).
		WillReturnResult(
			sqlmock.NewResult(0, 1),
		)

	mock.ExpectCommit()

	// Put something in Redis so we can verify invalidation.
	require.NoError(
		t,
		redisServer.Set("task:10", `{"id":10}`),
	)

	err := service.Delete(ctx, 10)

	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
	assert.False(t, redisServer.Exists("task:10"))
}

func TestTaskService_Delete_NoRowsAffected(t *testing.T) {
	service, mock, _ := testTaskService(t)

	ctx := context.WithValue(
		context.Background(),
		constants.UserIdKey,
		float64(99),
	)

	mock.ExpectQuery(
		`SELECT \* FROM "tasks" WHERE id = \$1`,
	).
		WithArgs(10, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
			}).AddRow(10),
		)

	mock.ExpectBegin()

	mock.ExpectExec(`UPDATE "tasks"`).
		WillReturnResult(
			sqlmock.NewResult(0, 0),
		)

	mock.ExpectRollback()

	err := service.Delete(ctx, 10)

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

func TestTaskService_GetByFilter_DefaultPagination(t *testing.T) {
	service, mock, _ := testTaskService(t)

	mock.ExpectQuery(
		`SELECT count\(\*\) FROM "tasks"`,
	).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(25),
		)

	mock.ExpectQuery(
		`SELECT \* FROM "tasks".*ORDER BY created_at DESC.*LIMIT \$1`,
	).
		WithArgs(20).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
			}).
				AddRow(
					1,
					"Task 1",
					"Description",
					models.TaskStatusPending,
					nil,
				),
		)

	result, err := service.GetByFilter(
		context.Background(),
		&dto.TaskListRequest{
			PaginationRequest: dto.PaginationRequest{
				Page:     0,
				PageSize: 0,
			},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Equal(t, int64(25), result.Total)
	assert.Equal(t, 2, result.TotalPages)
	assert.Len(t, result.Items, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_GetByFilter_Pagination(t *testing.T) {
	service, mock, _ := testTaskService(t)

	mock.ExpectQuery(
		`SELECT count\(\*\) FROM "tasks"`,
	).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(45),
		)

	mock.ExpectQuery(
		`SELECT \* FROM "tasks".* WHERE "tasks"."deleted_at" IS NULL ORDER BY created_at DESC.*LIMIT \$1 OFFSET \$2`,
	).
		WithArgs(20, 20).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
			}),
		)

	result, err := service.GetByFilter(
		context.Background(),
		&dto.TaskListRequest{
			PaginationRequest: dto.PaginationRequest{
				Page:     2,
				PageSize: 20,
			},
		},
	)

	require.NoError(t, err)

	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 20, result.PageSize)
	assert.Equal(t, int64(45), result.Total)
	assert.Equal(t, 3, result.TotalPages)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_GetByFilter_StatusFilter(t *testing.T) {
	service, mock, _ := testTaskService(t)

	status := models.TaskStatusCompleted

	mock.ExpectQuery(
		`SELECT count\(\*\) FROM "tasks" WHERE status = \$1`,
	).
		WithArgs(status).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(3),
		)

	mock.ExpectQuery(
		`SELECT \* FROM "tasks" WHERE status = \$1 AND "tasks"."deleted_at" IS NULL ORDER BY created_at DESC`,
	).
		WithArgs(status, 20).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
			}),
		)

	result, err := service.GetByFilter(
		context.Background(),
		&dto.TaskListRequest{
			Status: &status,
		},
	)

	require.NoError(t, err)

	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 1, result.TotalPages)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_GetByFilter_AssigneeFilter(t *testing.T) {
	service, mock, _ := testTaskService(t)

	assigneeID := 5

	mock.ExpectQuery(
		`SELECT count\(\*\) FROM "tasks" WHERE assignee_id = \$1`,
	).
		WithArgs(assigneeID).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(4),
		)

	mock.ExpectQuery(
		`SELECT \* FROM "tasks" WHERE assignee_id = \$1 AND "tasks"."deleted_at" IS NULL ORDER BY created_at DESC`,
	).
		WithArgs(assigneeID, 20).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"title",
				"description",
				"status",
				"assignee_id",
			}),
		)

	result, err := service.GetByFilter(
		context.Background(),
		&dto.TaskListRequest{
			AssigneeID: &assigneeID,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(4), result.Total)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_GetByFilter_CountError(t *testing.T) {
	service, mock, _ := testTaskService(t)

	dbError := errors.New("count failed")

	mock.ExpectQuery(
		`SELECT count\(\*\) FROM "tasks"`,
	).
		WillReturnError(dbError)

	result, err := service.GetByFilter(
		context.Background(),
		&dto.TaskListRequest{},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTaskService_GetByFilter_FindError(t *testing.T) {
	service, mock, _ := testTaskService(t)

	dbError := errors.New("find failed")

	mock.ExpectQuery(
		`SELECT count\(\*\) FROM "tasks"`,
	).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(10),
		)

	mock.ExpectQuery(
		`SELECT \* FROM "tasks".*ORDER BY created_at DESC`,
	).
		WillReturnError(dbError)

	result, err := service.GetByFilter(
		context.Background(),
		&dto.TaskListRequest{},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}
