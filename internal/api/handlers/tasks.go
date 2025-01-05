package handlers

import (
	"GoBlast/internal/api/middleware"
	"GoBlast/internal/tasks"
	"GoBlast/pkg/logger"
	"GoBlast/pkg/metrics"
	"GoBlast/pkg/queue"
	"GoBlast/pkg/response"
	"GoBlast/pkg/storage/models"
	"fmt"
	"time"

	"go.uber.org/zap"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Content struct {
	Type     string `json:"type" binding:"required"` // text, photo, video, etc.
	Text     string `json:"text,omitempty"`
	MediaURL string `json:"media_url,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type TaskRequest struct {
	Recipients []int64 `json:"recipients" binding:"required"` // Telegram Chat IDs
	Content    Content `json:"content" binding:"required"`
	Priority   string  `json:"priority,omitempty"` // high, medium, low
	Schedule   string  `json:"schedule,omitempty"` // RFC3339
}

type TaskNATSMessage struct {
	TaskID     string  `json:"task_id"`
	UserID     uint    `json:"user_id"`
	Recipients []int64 `json:"recipients"`
	Content    Content `json:"content"`
	Priority   string  `json:"priority,omitempty"`
	// Schedule  string   `json:"schedule,omitempty"` // если нужно
}

// TaskHandler обрабатывает задачи
type TaskHandler struct {
	repo       *tasks.TasksRepository
	natsClient *queue.NATSClient
}

// NewTaskHandler создаёт новый TaskHandler
func NewTaskHandler(repo *tasks.TasksRepository, natsClient *queue.NATSClient) *TaskHandler {
	return &TaskHandler{repo: repo, natsClient: natsClient}
}

func validateContent(content Content) error {
	switch content.Type {
	case "text":
		if content.Text == "" {
			return fmt.Errorf("text is required for type 'text'")
		}
	case "photo", "video", "animation":
		if content.MediaURL == "" {
			return fmt.Errorf("media_url is required for type '%s'", content.Type)
		}
	default:
		return fmt.Errorf("invalid content type: %s", content.Type)
	}
	return nil
}

func validatePriority(priority string) error {
	validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
	if priority != "" && !validPriorities[priority] {
		return fmt.Errorf("invalid priority value: %s", priority)
	}
	return nil
}

func parseSchedule(schedule string) (*time.Time, error) {
	if schedule == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, schedule)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// CreateTask Создаёт новую задачу
// @Summary Создать задачу
// @Description Создаёт новую задачу для отправки сообщений через Telegram.
// @Tags Tasks
// @Security BearerAuth
// @securityDefinitions.apikey BearerAuth
// @Accept json
// @Produce json
// @Param task body TaskRequest true "Создание задачи"
// @Success 201 {object} response.APIResponse "Задача успешно создана"
// @Failure 400 {object} response.APIResponse "Некорректные входные данные"
// @Failure 401 {object} response.APIResponse "Неавторизованный доступ"
// @Failure 500 {object} response.APIResponse "Внутренняя ошибка сервера"
// @Router /tasks [post]
// @example Request:
// /*
//
//	{
//	  "recipients": [123456789, 987654321],
//	  "content": {
//	    "type": "text",
//	    "text": "Привет! Это тестовое сообщение."
//	  },
//	  "priority": "high",
//	  "schedule": "2025-01-05T10:00:00Z"
//	}
//
// */
// @example Response (201):
// /*
//
//	{
//	  "success": true,
//	  "data": {
//	    "task_id": "a804bd98-8e4d-4e8d-9678-7e28b7a8408f",
//	    "status": "scheduled"
//	  }
//	}
//
// */
func (h *TaskHandler) CreateTask(c *gin.Context) {
	start := time.Now()
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("Ошибка привязки JSON", zap.Error(err))
		metrics.TaskFailedCounter.Inc()
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request payload"))
		return
	}

	// Извлечение user_id из контекста
	claims, exists := c.Get("claims")
	if !exists {
		logger.Log.Error("Ошибка авторизации: claims отсутствуют в контексте")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized"))
		return
	}

	userClaims, ok := claims.(*middleware.Claims)
	if !ok {
		logger.Log.Error("Ошибка авторизации: claims неверного формата")
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized"))
		return
	}
	userID := userClaims.UserID

	// Проверяем тип контента
	if err := validateContent(req.Content); err != nil {
		logger.Log.Error("Ошибка валидации контента", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse(err.Error()))
		return
	}

	// Проверяем Priority
	if err := validatePriority(req.Priority); err != nil {
		logger.Log.Error("Ошибка валидации приоритета", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse(err.Error()))
		return
	}

	// Генерация ID задачи
	taskID := uuid.New().String()

	// Парсим schedule
	schedule, err := parseSchedule(req.Schedule)
	if err != nil {
		logger.Log.Error("Ошибка парсинга расписания", zap.Error(err))
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid schedule format"))
		return
	}

	// Сериализуем Content в JSON (для БД)
	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		logger.Log.Error("Ошибка сериализации контента", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to serialize content"))
		return
	}

	// Создаём модель задачи
	task := &models.Task{
		ID:          taskID,
		UserID:      userID,
		MessageType: req.Content.Type,
		Content:     string(contentJSON),
		Priority:    req.Priority,
		Schedule:    schedule,
		Status:      "scheduled",
	}

	// Сохраняем в БД
	if err := h.repo.SaveTask(task); err != nil {
		logger.Log.Error("Ошибка сохранения задачи в БД", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to save task"))
		return
	}

	// Формируем сообщение для NATS
	msg := TaskNATSMessage{
		TaskID:     taskID,
		UserID:     userID,
		Recipients: req.Recipients,
		Content:    req.Content,
		Priority:   req.Priority,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Error("Ошибка сериализации сообщения для NATS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to marshal NATS message"))
		return
	}

	// Публикуем в NATS
	if err := h.natsClient.Conn.Publish("tasks.create", payload); err != nil {
		logger.Log.Error("Ошибка публикации в NATS", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to publish to NATS"))
		return
	}

	logger.Log.Info("Задача успешно создана",
		zap.String("task_id", taskID),
		zap.String("user_id", fmt.Sprintf("%d", userID)),
		zap.String("priority", req.Priority),
	)

	metrics.TaskCreatedCounter.Inc()
	metrics.TaskProcessingDuration.WithLabelValues("telegram").Observe(time.Since(start).Seconds())

	// Возвращаем результат
	c.JSON(http.StatusCreated, response.SuccessResponse(map[string]interface{}{
		"task_id": taskID,
		"status":  "scheduled",
	}))
}

// GetTask Возвращает задачу по ID
// @Summary Получить задачу
// @Description Возвращает детали задачи по её ID
// @securityDefinitions.apikey BearerAuth
// @Tags Tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Success 200 {object} response.APIResponse{data=models.Task} "Детали задачи"
// @Failure 401 {object} response.APIResponse "Неавторизованный доступ"
// @Failure 404 {object} response.APIResponse "Задача не найдена"
// @Failure 500 {object} response.APIResponse "Внутренняя ошибка сервера"
// @Router /tasks/{id} [get]
// @example Request:
// GET /tasks/a804bd98-8e4d-4e8d-9678-7e28b7a8408f
// @example Response (200):
//
//	{
//	  "success": true,
//	  "data": {
//	    "id": "a804bd98-8e4d-4e8d-9678-7e28b7a8408f",
//	    "user_id": 5,
//	    "message_type": "text",
//	    "content": "{\"type\":\"text\",\"text\":\"Привет! Это тестовое сообщение.\"}",
//	    "priority": "high",
//	    "schedule": "2025-01-05T10:00:00Z",
//	    "status": "scheduled",
//	    "created_at": "2025-01-04T21:37:39Z",
//	    "updated_at": "2025-01-04T21:37:39Z"
//	  }
//	}
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.repo.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse("Task not found"))
		return
	}
	c.JSON(http.StatusOK, response.SuccessResponse(task))
}
