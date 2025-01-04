package handlers

import (
	"GoBlast/api/middleware"
	"GoBlast/internal/tasks"
	"GoBlast/pkg/queue"
	"GoBlast/pkg/response"
	"GoBlast/pkg/storage/models"

	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

// Content описывает содержимое задачи
type Content struct {
	Type     string `json:"type" binding:"required"` // text, photo, video, etc.
	Text     string `json:"text,omitempty"`
	MediaURL string `json:"media_url,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

// TaskRequest представляет запрос на создание задачи
type TaskRequest struct {
	Recipients []int64 `json:"recipients" binding:"required"` // Telegram Chat IDs
	Content    Content `json:"content" binding:"required"`
	Priority   string  `json:"priority,omitempty"` // high, medium, low
	Schedule   string  `json:"schedule,omitempty"` // RFC3339
}

// TaskNATSMessage представляет сообщение для NATS
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

// CreateTask Создаёт новую задачу
// @Summary Создать задачу
// @Description Создаёт новую задачу для отправки сообщений через Telegram.
// @Tags Tasks
// @Security BearerAuth
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
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("JSON Bind Error: %v\n", err)
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request payload"))
		return
	}

	// Извлечение user_id из контекста
	claimsInterface, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized: Missing claims in context"))
		return
	}

	claims, ok := claimsInterface.(*middleware.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse("Unauthorized: Invalid claims format"))
		return
	}

	userID := claims.UserID

	// Проверяем тип контента
	switch req.Content.Type {
	case "text":
		if req.Content.Text == "" {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Text required for type='text'"))
			return
		}
	case "photo", "video", "animation":
		if req.Content.MediaURL == "" {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Media URL required for this type"))
			return
		}
	default:
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid content type"))
		return
	}

	// Генерация ID задачи
	taskID := uuid.New().String()

	// Парсим schedule
	var schedule *time.Time
	if req.Schedule != "" {
		parsed, err := time.Parse(time.RFC3339, req.Schedule)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid schedule format"))
			return
		}
		schedule = &parsed
	}

	// Сериализуем Content в JSON (хранить в БД)
	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to serialize content"))
		return
	}

	// Создаём модель задачи
	task := &models.Task{
		ID:          taskID,
		UserID:      userID, // Используем userID из claims
		MessageType: req.Content.Type,
		Content:     string(contentJSON), // для БД
		Priority:    req.Priority,
		Schedule:    schedule,
		Status:      "scheduled",
	}

	// Сохраняем в БД
	if err := h.repo.SaveTask(task); err != nil {
		log.Printf("Error saving task: %v\n", err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to save task"))
		return
	}

	// Формируем сообщение для NATS
	msg := TaskNATSMessage{
		TaskID:     taskID,
		UserID:     userID, // Используем userID из claims
		Recipients: req.Recipients,
		Content:    req.Content,
		Priority:   req.Priority,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to marshal NATS message"))
		return
	}

	// Публикуем в NATS
	if err := h.natsClient.Conn.Publish("tasks.create", payload); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to publish to NATS"))
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusCreated, response.SuccessResponse(map[string]interface{}{
		"task_id": taskID,
		"status":  "scheduled",
	}))
}

// GetTask Возвращает задачу по ID
// @Summary Получить задачу
// @Description Возвращает детали задачи по её ID
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
