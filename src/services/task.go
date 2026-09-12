package services

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/data/cache"
	"github.com/mahdipeydai/taskmanager-go/data/models"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/pkg/service_errors"
	"gorm.io/gorm"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100

	taskCacheKeyPrefix = "task:"
	taskCacheDuration  = 10 * time.Minute
)

type TaskService struct {
	logger logging.LoggerInterface
	cfg    *config.Config
	db     *gorm.DB
	redis  *redis.Client
}

func GetTaskService(database *gorm.DB, redis *redis.Client, cfg *config.Config, logger logging.LoggerInterface) *TaskService {
	return &TaskService{db: database, redis: redis, logger: logger, cfg: cfg}
}

func (s *TaskService) CreateTask(ctx context.Context, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssigneeID:  req.AssigneeID,
	}

	if task.Status == "" {
		task.Status = models.TaskStatusPending
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	err := tx.Create(task).Error
	if err != nil {
		tx.Rollback()
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		s.logger.Error(logging.Postgres, logging.Insert, "Failed to create record", extras)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.GetByID(ctx, task.Id)
}

func (s *TaskService) GetByID(ctx context.Context, id int) (*dto.TaskResponse, error) {
	cacheKey := s.getCacheKey(id)

	// Cache
	task, err := cache.Get[dto.TaskResponse](s.redis, cacheKey)
	if err == nil {
		return &task, nil
	}

	var model models.Task

	err = s.db.WithContext(ctx).
		Where("id = ? AND deleted_by is null", id).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, service_errors.ServiceError{
			EndUserMessage: service_errors.RecordNotFound,
		}
	}

	if err != nil {
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}

		s.logger.Error(
			logging.Postgres,
			logging.Select,
			"Failed to get task",
			extras,
		)

		return nil, err
	}

	response := s.toResponse(&model)

	if err := cache.Set(
		s.redis,
		cacheKey,
		*response,
		taskCacheDuration,
	); err != nil {
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}

		s.logger.Error(
			logging.Redis,
			logging.Insert,
			"Failed to cache task",
			extras,
		)
	}

	return response, nil
}

func (s *TaskService) Update(ctx context.Context, id int, req *dto.UpdateTaskRequest) (*dto.TaskResponse, error) {
	updateMap := make(map[string]interface{})

	if req.Title != nil {
		updateMap["title"] = *req.Title
	}

	if req.Description != nil {
		updateMap["description"] = *req.Description
	}

	if req.Status != nil {
		updateMap["status"] = *req.Status
	}

	if req.AssigneeID != nil {
		updateMap["assignee_id"] = *req.AssigneeID
	}

	updateMap["updated_at"] = time.Now().UTC()
	updateMap["updated_by"] = int64(ctx.Value(constants.UserIdKey).(float64))

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var task models.Task

	if err := tx.
		Where("id = ? AND deleted_by is null", id).
		First(&task).Error; err != nil {

		tx.Rollback()

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service_errors.ServiceError{
				EndUserMessage: service_errors.RecordNotFound,
			}
		}

		return nil, err
	}

	if err := tx.
		Model(&models.Task{}).
		Where("id = ? AND deleted_by is null", id).
		Updates(updateMap).Error; err != nil {

		tx.Rollback()

		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}

		s.logger.Error(
			logging.Postgres,
			logging.Update,
			"Failed to update task",
			extras,
		)

		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Invalidate cached task.
	if err := cache.Del(s.redis, s.getCacheKey(id)); err != nil {
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}

		s.logger.Error(
			logging.Redis,
			logging.Delete,
			"Failed to invalidate task cache",
			extras,
		)
	}

	return s.GetByID(ctx, id)
}

func (s *TaskService) Delete(ctx context.Context, id int) error {
	var task models.Task

	err := s.db.WithContext(ctx).
		Where("id = ?", id).
		First(&task).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return service_errors.ServiceError{
			EndUserMessage: service_errors.RecordNotFound,
		}
	}

	if err != nil {
		return err
	}

	deleteMap := map[string]interface{}{
		"deleted_by": sql.NullInt64{Int64: int64(ctx.Value(constants.UserIdKey).(float64)), Valid: true},
		"deleted_at": time.Now().UTC(),
	}
	deleteMap["updated_at"] = time.Now().UTC()
	deleteMap["updated_by"] = int64(ctx.Value(constants.UserIdKey).(float64))

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	model := new(models.Task)
	count := tx.Model(model).Where("id = ? AND deleted_by is null", id).Updates(deleteMap).RowsAffected
	if count == 0 {
		tx.Rollback()
		err := service_errors.ServiceError{EndUserMessage: service_errors.RecordNotFound}
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		s.logger.Error(logging.Postgres, logging.Update, "Failed to delete record", extras)
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	if err := cache.Del(s.redis, s.getCacheKey(id)); err != nil {
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}

		s.logger.Error(
			logging.Redis,
			logging.Delete,
			"Failed to invalidate task cache",
			extras,
		)
	}

	return nil
}

func (s *TaskService) GetByFilter(ctx context.Context, req *dto.TaskListRequest) (*dto.TaskListResponse, error) {
	page := req.Page
	if page < 1 {
		page = defaultPage
	}

	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = defaultPageSize
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	query := s.db.WithContext(ctx).
		Model(&models.Task{})

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if req.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *req.AssigneeID)
	}

	var totalRows int64

	if err := query.Count(&totalRows).Error; err != nil {
		return nil, err
	}

	var tasks []models.Task

	offset := (page - 1) * pageSize

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	items := make([]dto.TaskResponse, 0, len(tasks))

	for _, task := range tasks {
		items = append(items, *s.toResponse(&task))
	}

	totalPages := int(math.Ceil(
		float64(totalRows) / float64(pageSize),
	))

	return &dto.TaskListResponse{
		PaginationResponse: dto.PaginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      totalRows,
			TotalPages: totalPages,
		},
		Items: items,
	}, nil
}

func (s *TaskService) toResponse(task *models.Task) *dto.TaskResponse {
	return &dto.TaskResponse{
		ID:          task.Id,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		AssigneeID:  task.AssigneeID,
	}
}

func (s *TaskService) getCacheKey(id int) string {
	return taskCacheKeyPrefix + strconv.Itoa(id)
}
