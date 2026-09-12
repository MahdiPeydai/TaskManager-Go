package models

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
)

type Task struct {
	BaseModel
	Title       string     `gorm:"size:255;not null"`
	Description string     `gorm:"type:text"`
	Status      TaskStatus `gorm:"type:varchar(20);not null;default:'pending';index"`
	AssigneeID  *int       `gorm:"index"`
	Assignee    *User      `gorm:"foreignKey:AssigneeID"`
}
