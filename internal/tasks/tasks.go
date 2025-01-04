package tasks

import (
	"GoBlast/pkg/storage/models"
	"errors"

	"gorm.io/gorm"
)

type TasksRepository struct {
	db *gorm.DB
}

func NewTasksRepository(db *gorm.DB) *TasksRepository {
	return &TasksRepository{db: db}
}

// SaveTask сохраняет задачу
func (r *TasksRepository) SaveTask(task *models.Task) error {

	return r.db.Create(task).Error
}

// GetTaskByID возвращает задачу по ID
func (r *TasksRepository) GetTaskByID(id string) (*models.Task, error) {
	var task models.Task
	err := r.db.First(&task, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("task not found")
	}

	return &task, err
}

func (r *TasksRepository) UpdateStatus(taskID string, newStatus string) error {

	var t models.Task
	if err := r.db.First(&t, "id = ?", taskID).Error; err != nil {
		return err
	}

	// Меняем статус
	t.Status = newStatus

	// Сохраняем
	return r.db.Save(&t).Error
}
