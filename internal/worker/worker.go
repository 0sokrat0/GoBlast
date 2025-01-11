package worker

import (
	"GoBlast/pkg/logger"
	"GoBlast/pkg/storage/models"
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	tele "gopkg.in/telebot.v4"
	"strings"
	"sync"
	"time"
)

// WorkerRepo — интерфейс для репозитория, чтобы обновлять статус и статистику в БД,
// а также, если нужно, публиковать событие о завершении задачи.
type WorkerRepo interface {
	UpdateStatusAndStats(taskID, newStatus string, stats models.Stats) error
	PublishCompleteStatus(taskID string, finalStats models.Stats) error
}

// BotInterface — упрощённый интерфейс телеграм-бота (для тестирования).
type BotInterface interface {
	Send(to tele.Recipient, what interface{}, options ...interface{}) (*tele.Message, error)
}

// TaskItem описывает один «подзадачу» (конкретному получателю).
type TaskItem struct {
	TaskID    string
	Recipient int64
	Content   Content
}

// Worker отвечает за рассылку сообщений от имени одного бота (botToken).
type Worker struct {
	Bot         BotInterface
	WG          sync.WaitGroup
	RateLimiter *rate.Limiter
	TaskChan    chan TaskItem
	NumWorkers  int
	Repo        WorkerRepo

	mu    sync.Mutex
	stats map[string]*models.Stats // key=TaskID -> накопленная статистика
}

// NewWorker создаёт воркер с начальным rate-limit (по умолчанию 10 msg/sec).
func NewWorker(botToken string, numWorkers int, repo WorkerRepo) (*Worker, error) {
	logger.Log.Info("[Worker] Инициализация воркера",
		zap.String("bot_token", botToken),
		zap.Int("num_workers", numWorkers))

	bot, err := tele.NewBot(tele.Settings{
		Token:     botToken,
		Poller:    nil,
		ParseMode: "HTML",
	})
	if err != nil {
		logger.Log.Error("Ошибка создания бота", zap.Error(err))
		return nil, err
	}

	w := &Worker{
		Bot:         bot,
		RateLimiter: rate.NewLimiter(rate.Every(time.Second/10), 1), // medium
		TaskChan:    make(chan TaskItem),
		NumWorkers:  numWorkers,
		Repo:        repo,
		stats:       make(map[string]*models.Stats),
	}
	return w, nil
}

// Start запускает N горутин (workerLoop), каждая читает из TaskChan и обрабатывает сообщения.
func (w *Worker) Start() {
	logger.Log.Info("[Worker] Запуск воркера", zap.Int("num_workers", w.NumWorkers))
	for i := 0; i < w.NumWorkers; i++ {
		w.WG.Add(1)
		go w.workerLoop(i)
	}
}

// AddTask выставляет приоритет (меняет RateLimiter), заводит/дополняет статистику
// и выкладывает всех получателей (TaskItem) в канал.
func (w *Worker) AddTask(task TaskNATSMessage) {
	logger.Log.Info("[Worker] Получена задача",
		zap.String("task_id", task.TaskID),
		zap.Int("recipients_count", len(task.Recipients)),
		zap.String("priority", task.Priority))

	// Настраиваем rate-limit в зависимости от приоритета
	w.mu.Lock()
	switch strings.ToLower(task.Priority) {
	case "high":
		// 30 msg/sec
		w.RateLimiter = rate.NewLimiter(rate.Every(time.Second/30), 1)
		logger.Log.Info("Установлен высокий приоритет (30 msg/sec)")
	case "low":
		// 2 msg/sec
		w.RateLimiter = rate.NewLimiter(rate.Every(time.Second/2), 1)
		logger.Log.Info("Установлен низкий приоритет (2 msg/sec)")
	default:
		// По умолчанию medium (10 msg/sec)
		w.RateLimiter = rate.NewLimiter(rate.Every(time.Second/10), 1)
		logger.Log.Info("Установлен средний приоритет (10 msg/sec)")
	}

	// Заводим/получаем статистику для данного TaskID
	st, exists := w.stats[task.TaskID]
	if !exists {
		st = &models.Stats{
			ByContentType: make(map[string]int64),
			ErrorCounts:   make(map[string]int64),
			StartTime:     time.Now(),
		}
		w.stats[task.TaskID] = st
	}
	// Увеличиваем ExpectedCount
	st.ExpectedCount += int64(len(task.Recipients))
	w.mu.Unlock()

	// Выкладываем всех получателей в канал
	for _, recipient := range task.Recipients {
		w.TaskChan <- TaskItem{
			TaskID:    task.TaskID,
			Recipient: recipient,
			Content:   task.Content,
		}
	}
}

// workerLoop читает из TaskChan, соблюдает RateLimiter, отправляет сообщение
// и при успехе/ошибке инкрементирует статистику (Sent/Failed).
func (w *Worker) workerLoop(workerID int) {
	defer w.WG.Done()
	logger.Log.Info("[Worker] workerLoop запущен",
		zap.Int("worker_id", workerID))

	for item := range w.TaskChan {
		logger.Log.Info("[Worker] Обработка получателя",
			zap.Int("worker_id", workerID),
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient),
			zap.String("content_type", item.Content.Type))

		// Rate-limit
		if err := w.RateLimiter.Wait(context.Background()); err != nil {
			logger.Log.Error("[Worker] Ошибка rate-limiter",
				zap.Int("worker_id", workerID),
				zap.Error(err))
			w.incrementFailed(item.TaskID, item.Content.Type, err)
			continue
		}

		// Попытка отправки
		if err := w.sendMessage(item); err != nil {
			// В sendMessage(...) при ошибке вызывается handleTgError(...), которая делает incrementFailed
			// Здесь просто переходим к следующему
			continue
		}

		// Успешная отправка
		logger.Log.Info("[Worker] Успешно отправлено",
			zap.Int("worker_id", workerID),
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient),
			zap.String("content_type", item.Content.Type))
		w.incrementSent(item.TaskID, item.Content.Type)
	}

	logger.Log.Info("[Worker] workerLoop завершается", zap.Int("worker_id", workerID))
}

// sendMessage — единая точка для отправки сообщения любым способом.
func (w *Worker) sendMessage(item TaskItem) error {
	c := item.Content
	logger.Log.Info("[Worker] sendMessage",
		zap.String("task_id", item.TaskID),
		zap.Int64("recipient", item.Recipient),
		zap.String("type", c.Type),
		zap.String("media_id", c.MediaID),
		zap.String("media_url", c.MediaURL))

	var err error
	switch c.Type {
	case "text":
		_, err = w.Bot.Send(tele.ChatID(item.Recipient), c.Text)

	case "photo":
		err = w.sendPhoto(item)

	case "animation":
		err = w.sendAnimation(item)

	case "video":
		err = w.sendVideo(item)

	case "document":
		err = w.sendDocument(item)

	case "audio":
		err = w.sendAudio(item)

	case "circle":
		err = w.sendCircle(item)

	default:
		err = fmt.Errorf("неподдерживаемый тип контента: %s", c.Type)
	}

	return w.handleTgError(item, err)
}

// incrementSent — при успехе
func (w *Worker) incrementSent(taskID, contentType string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	st := w.stats[taskID]
	if st == nil {
		return
	}
	st.TotalSent++
	st.ProcessedCount++
	st.ByContentType[contentType]++

	if st.ProcessedCount == st.ExpectedCount {
		w.finishTask(taskID, st)
	}
}

// incrementFailed — при ошибке
func (w *Worker) incrementFailed(taskID, contentType string, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	st := w.stats[taskID]
	if st == nil {
		return
	}
	st.TotalFailed++
	st.ProcessedCount++

	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "chat not found") {
			st.ErrorCounts["NOT_FOUND"]++
		} else if strings.Contains(msg, "FLOOD_WAIT") {
			st.ErrorCounts["FLOOD_WAIT"]++
		} else {
			st.ErrorCounts["other"]++
		}
	}

	if st.ProcessedCount == st.ExpectedCount {
		w.finishTask(taskID, st)
	}
}

// finishTask — когда ProcessedCount == ExpectedCount, задача завершается
func (w *Worker) finishTask(taskID string, finalStats *models.Stats) {
	logger.Log.Info("[Worker] Задача завершена",
		zap.String("task_id", taskID))

	elapsed := time.Since(finalStats.StartTime).Seconds()
	finalStats.TimeSpent = elapsed

	// 1. Обновляем статус и статистику в БД
	if err := w.Repo.UpdateStatusAndStats(taskID, "complete", *finalStats); err != nil {
		logger.Log.Error("[Worker] Ошибка UpdateStatusAndStats",
			zap.String("task_id", taskID),
			zap.Error(err))
	}

	// 2. Публикация события (если нужно)
	if err := w.Repo.PublishCompleteStatus(taskID, *finalStats); err != nil {
		logger.Log.Error("[Worker] Ошибка PublishCompleteStatus",
			zap.String("task_id", taskID),
			zap.Error(err))
	}

	// 3. Лог для отладки
	logger.Log.Info("finishTask() debug",
		zap.String("task_id", taskID),
		zap.Int64("ExpectedCount", finalStats.ExpectedCount),
		zap.Int64("ProcessedCount", finalStats.ProcessedCount))

	// 4. Удаляем запись из stats
	delete(w.stats, taskID)
}
