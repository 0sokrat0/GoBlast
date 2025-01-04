package db

import (
	"GoBlast/pkg/logger"
	"GoBlast/pkg/storage/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(
		&models.AuthUser{},
		&models.Task{},
	)
	if err != nil {
		return err
	}
	logger.Log.Info("Миграция прошла успешно")
	return nil
}
