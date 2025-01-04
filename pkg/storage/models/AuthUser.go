// pkg/storage/models/auth_user.go

package models

import (
	"time"
)

type AuthUser struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`  // Уникальное имя пользователя
	Token     string    `gorm:"type:varchar(512)" json:"token"`   // Зашифрованный токен (увеличили длину)
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"` // Время создания записи
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"` // Время последнего обновления
}
