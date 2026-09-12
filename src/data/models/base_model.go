package models

import (
	"time"

	"github.com/mahdipeydai/taskmanager-go/constants"
	"gorm.io/gorm"
)

type BaseModel struct {
	Id int `gorm:"primaryKey;column:id;auto_increment"`

	CreatedAt time.Time      `gorm:"column:created_at;auto_now_add;type:TIMESTAMP with time zone"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:TIMESTAMP with time zone; null"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	CreatedBy int64  `gorm:"column:created_by;not null"`
	UpdatedBy *int64 `gorm:"column:updated_by;null"`
	DeletedBy *int64 `gorm:"column:deleted_by;null"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	var userId int64 = -1
	if val := tx.Statement.Context.Value(constants.UserIdKey); val != nil {
		if userIdFloat, ok := val.(float64); ok {
			userId = int64(userIdFloat)
		}
	}
	m.CreatedAt = time.Now().UTC()
	m.CreatedBy = userId
	return nil
}
