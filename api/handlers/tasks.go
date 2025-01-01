package handlers

import (
	"GoBlast/internal/tasks"
	"GoBlast/pkg/queue"
	"GoBlast/pkg/response"
	"GoBlast/pkg/storage/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type Content struct {
	Type     string `json:"type" binding:"required"` // Тип сообщения: text, photo, video, etc.
	Text     string `json:"text,omitempty"`          // Текст сообщения (опционально)
	MediaURL string `json:"media_url,omitempty"`     // URL медиа (фото, видео, анимация)
	Caption  string `json:"caption,omitempty"`       // Подпись к медиа (опционально)
}

type TaskRequest struct {
	UserID     uint    `json:"user_id" binding:"required"`    // ID пользователя, создавшего задачу
	Recipients []int64 `json:"recipients" binding:"required"` // Список ID получателей
	Content    Content `json:"content" binding:"required"`    // Содержимое сообщения
	Priority   string  `json:"priority,omitempty"`            // Приоритет: high, medium, low
	Schedule   string  `json:"schedule,omitempty"`            // Время отправки (ISO 8601)
}

type TaskHandler struct {
	repo       *tasks.TasksRepository
	natsClient *queue.NATSClient
}

func NewTaskHandler(repo *tasks.TasksRepository, natsClient *queue.NATSClient) *TaskHandler {
	return &TaskHandler{repo: repo, natsClient: natsClient}
}

// CreateTask обрабатывает создание новой задачи
// @Summary Создание новой задачи
// @Description Создаёт новую задачу для рассылки сообщений
// @Tags Tasks
// @Accept json
// @Produce json
// @Param input body TaskRequest true "Данные задачи"
// @Success 201 {object} map[string]interface{} "Успешный ответ с ID задачи"
// @Failure 400 {object} map[string]interface{} "Ошибка валидации запроса"
// @Failure 500 {object} map[string]interface{} "Ошибка сохранения задачи"
// @Security BearerAuth
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid request payload"))
		return
	}

	// Валидация приоритета
	if err := validator.New().Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid priority value"))
		return
	}

	// Проверка контента
	switch req.Content.Type {
	case "text":
		if req.Content.Text == "" {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Text is required for type 'text'"))
			return
		}
	case "photo", "video", "animation":
		if req.Content.MediaURL == "" {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Media URL is required for this type"))
			return
		}
	default:
		c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid content type"))
		return
	}

	// Генерация ID задачи
	taskID := uuid.New().String()

	// Парсинг времени расписания
	var schedule *time.Time
	if req.Schedule != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.Schedule)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorResponse("Invalid schedule format"))
			return
		}
		schedule = &parsedTime
	}

	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to serialize content"))
		return
	}

	task := &models.Task{
		ID:          taskID,
		UserID:      req.UserID,
		MessageType: req.Content.Type,
		Content:     string(contentJSON), // В зависимости от типа
		Priority:    req.Priority,
		Schedule:    schedule,
		Status:      "scheduled",
	}

	// Сохранение задачи
	if err := h.repo.SaveTask(task); err != nil {
		log.Printf("Error saving task: %v\n", err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to save task"))
		return
	}

	messagePayload, err := json.Marshal(task)
	// Или можно собрать только нужные поля без лишней информации
	if err != nil {
		log.Printf("Error marshaling task to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to publish task"))
		return
	}
	err = h.natsClient.Conn.Publish("tasks.create", messagePayload)
	if err != nil {
		log.Printf("Error publishing to NATS: %v", err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse("Failed to publish task"))
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse(map[string]interface{}{
		"task_id": taskID,
		"status":  "scheduled",
	}))
}

// GetTask обрабатывает получение задачи по ID
// @Summary Получение задачи по ID
// @Description Возвращает задачу по её ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Success 200 {object} models.Task "Успешный ответ с данными задачи"
// @Failure 404 {object} map[string]interface{} "Задача не найдена"
// @Security BearerAuth
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.repo.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse("Task not found"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse(task))
}
