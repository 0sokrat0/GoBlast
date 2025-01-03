package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	UserID      uint           `gorm:"not null"`
	MessageType string         `gorm:"type:varchar(20);not null"`
	Content     string         `gorm:"type:jsonb;not null"`
	Priority    string         `gorm:"type:varchar(10);default:'medium'"`
	Schedule    *time.Time     `gorm:"type:timestamp"`
	Status      string         `gorm:"type:varchar(20);not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
