package tasks

import (
	"GoBlast/pkg/logger"
	"GoBlast/pkg/queue"
	"GoBlast/pkg/storage/models"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TasksRepository struct {
	db         *gorm.DB
	natsClient *queue.NATSClient
}

func NewTasksRepository(db *gorm.DB) *TasksRepository {
	return &TasksRepository{db: db}
}

func (r *TasksRepository) SetNATSClient(nc *queue.NATSClient) {
	r.natsClient = nc
}

// SaveTask сохраняет задачу (если нужно).
func (r *TasksRepository) SaveTask(task *models.Task) error {
	return r.db.Create(task).Error
}

// GetTaskByID возвращает задачу по ID (если нужно).
func (r *TasksRepository) GetTaskByID(id string) (*models.Task, error) {
	var t models.Task
	if err := r.db.First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("задача %s не найдена", id)
		}
		return nil, err
	}
	return &t, nil
}

// UpdateStatusAndStats удовлетворяет интерфейсу worker.WorkerRepo.
// Принимает worker.Stats, сериализует в JSON, пишет в колонку `stats` таблицы tasks.
func (r *TasksRepository) UpdateStatusAndStats(taskID, newStatus string, stats models.Stats) error {
	statsBytes, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}
	statsStr := string(statsBytes)

	// Обновляем поля в БД
	if err := r.db.Model(&models.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status": newStatus,
			"stats":  statsStr,
		}).Error; err != nil {
		return err
	}

	logger.Log.Info("Обновлены статус и статистика задачи в БД",
		zap.String("task_id", taskID),
		zap.String("status", newStatus))

	// Если требуется опубликовать в NATS
	if newStatus == "complete" && r.natsClient != nil {
		completeMsg, _ := json.Marshal(map[string]interface{}{
			"task_id": taskID,
			"status":  newStatus,
			"stats":   stats,
		})
		if err := r.natsClient.Conn.Publish("tasks.complete", completeMsg); err != nil {
			logger.Log.Error("Ошибка публикации в tasks.complete", zap.Error(err))
		} else {
			logger.Log.Info("Опубликовано событие tasks.complete в NATS",
				zap.String("task_id", taskID))
		}
	}

	return nil
}

func (r *TasksRepository) PublishCompleteStatus(taskID string, finalStats models.Stats) error {
	if r.natsClient == nil {
		return nil
	}
	completeMsg, _ := json.Marshal(map[string]interface{}{
		"task_id": taskID,
		"status":  "complete",
		"stats":   finalStats,
	})
	return r.natsClient.Conn.Publish("tasks.complete", completeMsg)
}
