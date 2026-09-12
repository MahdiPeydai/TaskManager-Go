package dto

import "github.com/mahdipeydai/taskmanager-go/data/models"

type CreateTaskRequest struct {
	Title       string            `json:"title" binding:"required,min=1,max=255"`
	Description string            `json:"description" binding:"required"`
	Status      models.TaskStatus `json:"status" binding:"required"`
	AssigneeID  *int              `json:"assignee_id"`
}

type UpdateTaskRequest struct {
	Title       *string            `json:"title,omitempty" binding:"min=1,max=255"`
	Description *string            `json:"description,omitempty"`
	Status      *models.TaskStatus `json:"status,omitempty"`
	AssigneeID  *int               `json:"assignee_id,omitempty"`
}

type TaskResponse struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      models.TaskStatus `json:"status"`
	AssigneeID  *int              `json:"assignee_id"`
}

type TaskListRequest struct {
	PaginationRequest
	Status     *models.TaskStatus `form:"status"`
	AssigneeID *int               `form:"assignee_id"`
}

type TaskListResponse struct {
	Items []TaskResponse `json:"items"`
	PaginationResponse
}
