package worker

import (
	"GoBlast/internal/tasks"
	"GoBlast/pkg/logger"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BotManager struct {
	mu      sync.Mutex
	workers map[string]*Worker
}

// NewBotManager возвращает новый менеджер ботов
func NewBotManager() *BotManager {
	return &BotManager{
		workers: make(map[string]*Worker),
	}
}

func (bm *BotManager) StartTask(botToken string, natsMsg TaskNATSMessage, db *gorm.DB) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	worker, exists := bm.workers[botToken]
	if !exists {
		// Создаём repo
		repo := tasks.NewTasksRepository(db)

		// Создаём воркер
		w, err := NewWorker(botToken, 10, repo)
		if err != nil {
			logger.Log.Error("Ошибка создания воркера для бота", zap.Error(err))
			return
		}
		worker = w

		// Сохраняем
		bm.workers[botToken] = worker

		// Запускаем
		worker.Start()
	}

	// Добавляем задачу в воркер
	worker.AddTask(natsMsg)
}
