package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/api/helpers"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/data/cache"
	"github.com/mahdipeydai/taskmanager-go/data/db"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/services"
)

type TasksHandler struct {
	cfg         *config.Config
	taskService *services.TaskService
}

func GetTasksHandler(cfg *config.Config) *TasksHandler {
	logger := logging.GetLogger(cfg)

	return &TasksHandler{
		taskService: services.GetTaskService(db.GetDB(), cache.GetRedis(), cfg, logger),
		cfg:         cfg,
	}
}

// Create godoc
//
//	@Summary		Create Task
//	@Description	Create a new task.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Param			Request	body		dto.CreateTaskRequest	true	"Create task"
//	@Success		201		{object}	helpers.BaseHTTPResponse	"Success"
//	@Failure		400		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		404		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		500		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Router			/v1/tasks [post]
//	@Security		AuthBearer
func (th TasksHandler) Create(c *gin.Context) {
	req := new(dto.CreateTaskRequest)

	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithValidationErrors(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	userID := c.GetInt(constants.UserIdKey)
	roles := c.GetStringSlice(constants.RolesKey)

	task, err := th.taskService.CreateTask(c.Request.Context(), userID, roles, req)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	c.JSON(
		http.StatusCreated,
		helpers.GenerateBaseResponse(
			task,
			true,
			0,
		),
	)
}

// GetByID godoc
//
//	@Summary		Get Task
//	@Description	Get a task by ID.
//	@Tags			Tasks
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Success		200	{object}	helpers.BaseHTTPResponse	"Success"
//	@Failure		400	{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		404	{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		500	{object}	helpers.BaseHTTPResponse	"Failed"
//	@Router			/v1/tasks/{id} [get]
//	@Security		AuthBearer
func (th TasksHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			helpers.GenerateBaseResponseWithValidationErrors(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	userID := c.GetInt(constants.UserIdKey)
	roles := c.GetStringSlice(constants.RolesKey)

	task, err := th.taskService.GetByID(c.Request.Context(), userID, roles, id)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	c.JSON(
		http.StatusOK,
		helpers.GenerateBaseResponse(
			task,
			true,
			0,
		),
	)
}

// GetByFilter godoc
//
//	@Summary		Get Tasks
//	@Description	Get paginated tasks with optional status and assignee filters.
//	@Tags			Tasks
//	@Produce		json
//	@Param			page			query	int				false	"Page number"
//	@Param			page_size		query	int				false	"Number of items per page"
//	@Param			status			query	string			false	"Task status"
//	@Param			assignee_id		query	int				false	"Assignee ID"
//	@Success		200				{object}	helpers.BaseHTTPResponse	"Success"
//	@Failure		400				{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		500				{object}	helpers.BaseHTTPResponse	"Failed"
//	@Router			/v1/tasks [get]
//	@Security		AuthBearer
func (th TasksHandler) GetByFilter(c *gin.Context) {
	req := new(dto.TaskListRequest)

	err := c.ShouldBindQuery(req)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithValidationErrors(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	userID := c.GetInt(constants.UserIdKey)
	roles := c.GetStringSlice(constants.RolesKey)

	tasks, err := th.taskService.GetByFilter(c.Request.Context(), userID, roles, req)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	c.JSON(
		http.StatusOK,
		helpers.GenerateBaseResponse(
			tasks,
			true,
			0,
		),
	)
}

// Update godoc
//
//	@Summary		Update Task
//	@Description	Update an existing task.
//	@Tags			Tasks
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Task ID"
//	@Param			Request	body		dto.UpdateTaskRequest	true	"Update task"
//	@Success		200		{object}	helpers.BaseHTTPResponse	"Success"
//	@Failure		400		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		404		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		500		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Router			/v1/tasks/{id} [put]
//	@Security		AuthBearer
func (th TasksHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			helpers.GenerateBaseResponseWithValidationErrors(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	req := new(dto.UpdateTaskRequest)

	err = c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithValidationErrors(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	userID := c.GetInt(constants.UserIdKey)
	roles := c.GetStringSlice(constants.RolesKey)

	task, err := th.taskService.Update(
		c.Request.Context(),
		userID,
		roles,
		id,
		req,
	)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	c.JSON(
		http.StatusOK,
		helpers.GenerateBaseResponse(
			task,
			true,
			0,
		),
	)
}

// Delete godoc
//
//	@Summary		Delete Task
//	@Description	Soft delete a task.
//	@Tags			Tasks
//	@Produce		json
//	@Param			id	path		int	true	"Task ID"
//	@Success		200	{object}	helpers.BaseHTTPResponse	"Success"
//	@Failure		400	{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		404	{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		500	{object}	helpers.BaseHTTPResponse	"Failed"
//	@Router			/v1/tasks/{id} [delete]
//	@Security		AuthBearer
func (th TasksHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			helpers.GenerateBaseResponseWithValidationErrors(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	userID := c.GetInt(constants.UserIdKey)
	roles := c.GetStringSlice(constants.RolesKey)

	err = th.taskService.Delete(c.Request.Context(), userID, roles, id)
	if err != nil {
		c.AbortWithStatusJSON(
			helpers.TranslateErrorToStatusCode(err),
			helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-1,
				err,
			),
		)
		return
	}

	c.JSON(
		http.StatusOK,
		helpers.GenerateBaseResponse(
			nil,
			true,
			0,
		),
	)
}
