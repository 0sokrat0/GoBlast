package worker

import (
	"GoBlast/api/middleware"
	"GoBlast/configs"
	"GoBlast/internal/tasks"
	"GoBlast/internal/users"
	"GoBlast/pkg/encryption"
	"GoBlast/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"GoBlast/pkg/queue"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	tele "gopkg.in/telebot.v4"
	"log"
)

// TaskNATSMessage — структура сообщения, которое приходит из NATS
type TaskNATSMessage struct {
	TaskID     string  `json:"task_id"` // ID задачи
	UserID     uint    `json:"user_id"`
	Recipients []int64 `json:"recipients"`         // ID получателей
	Content    Content `json:"content"`            // контент
	Priority   string  `json:"priority,omitempty"` // приоритет
	// Schedule ... (добавьте, если нужно)
}

// Content — описание контента (тип, текст/медиа и т.д.)
type Content struct {
	Type     string `json:"type"`      // "text", "photo", "video", ...
	Text     string `json:"text"`      // если type="text"
	MediaURL string `json:"media_url"` // если фото, видео, документ...
	Caption  string `json:"caption"`   // подпись
}

func Worker(cfg *configs.Config, database *gorm.DB, natsClient *queue.NATSClient) error {

	_, err := natsClient.Conn.QueueSubscribe("tasks.create", "worker-group", func(msg *nats.Msg) {
		var natsMsg TaskNATSMessage
		if err := json.Unmarshal(msg.Data, &natsMsg); err != nil {
			logger.Log.Error("Failed to unmarshal NATS message:", zap.Error(err))
			return
		}
		log.Printf("Получено событие о задаче ID=%s, user=%d, contentType=%s", natsMsg.TaskID, natsMsg.UserID, natsMsg.Content.Type)
		tasksRepo := tasks.NewTasksRepository(database)
		if err := tasksRepo.UpdateStatus(natsMsg.TaskID, "in_progress"); err != nil {
			logger.Log.Error("Failed to set in_progress for task", zap.Error(err), zap.String("task_id", natsMsg.TaskID))
		}
		if err := processTask(database, natsMsg, cfg); err != nil {
			logger.Log.Error("Failed to process task", zap.Error(err), zap.String("task_id", natsMsg.TaskID))

			tasksRepo.UpdateStatus(natsMsg.TaskID, "failed")
			return
		}
		if err := tasksRepo.UpdateStatus(natsMsg.TaskID, "completed"); err != nil {
			logger.Log.Error("Failed to set completed for task", zap.Error(err), zap.String("task_id", natsMsg.TaskID))
		}
	})
	if err != nil {
		logger.Log.Error("неFailed to subscribe to tasks.create queue:", zap.Error(err))
	}
	return nil
}

func processTask(database *gorm.DB, natsMsg TaskNATSMessage, cfg *configs.Config) error {
	log.Printf("Начало обработки задачи ID=%s для пользователя ID=%d", natsMsg.TaskID, natsMsg.UserID)

	// 1) Ищем пользователя => бот-токен
	userRepo := users.NewAuthUserRepository(database)
	userData, err := userRepo.FindByID(natsMsg.UserID)
	if err != nil {
		return fmt.Errorf("error finding user by ID=%d: %v", natsMsg.UserID, err)
	}
	log.Printf("Пользователь найден: %s", userData.Username)

	// Декодирование base64 строки обратно в []byte
	encryptedToken, err := base64.StdEncoding.DecodeString(userData.Token)
	if err != nil {
		return fmt.Errorf("error decoding token for user %d: %v", natsMsg.UserID, err)
	}
	log.Printf("Токен пользователя декодирован")

	// Дешифрование токена
	decryptedTokenBytes, err := encryption.Decrypt(encryptedToken, []byte(middleware.EncryptionKey))
	if err != nil {
		return fmt.Errorf("error decrypting token for user %d: %v", natsMsg.UserID, err)
	}
	botToken := string(decryptedTokenBytes)
	if botToken == "" {
		return fmt.Errorf("user %d has empty bot token", natsMsg.UserID)
	}
	log.Printf("Токен пользователя дешифрован: %s", botToken)

	// 2) Создаём бота
	bot, err := tele.NewBot(tele.Settings{
		Token:     botToken,
		Poller:    nil,
		ParseMode: "HTML",
	})
	if err != nil {
		return fmt.Errorf("error creating bot: %v", err)
	}
	log.Printf("Бот успешно создан")

	// 3) Перебираем Recipients и отправляем
	content := natsMsg.Content
	for _, chatID := range natsMsg.Recipients {
		log.Printf("Отправка сообщения в чат ID=%d", chatID)
		switch content.Type {
		case "text":
			_, err = bot.Send(tele.ChatID(chatID), content.Text)
		case "photo":
			photo := &tele.Photo{File: tele.FromURL(content.MediaURL), Caption: content.Caption}
			_, err = bot.Send(tele.ChatID(chatID), photo)
		case "video":
			video := &tele.Video{File: tele.FromURL(content.MediaURL), Caption: content.Caption}
			_, err = bot.Send(tele.ChatID(chatID), video)
		// Добавьте "document", "animation", "voice" и т.д. по аналогии
		default:
			log.Printf("Unknown content type: %s", content.Type)
			continue
		}
		if err != nil {
			log.Printf("Failed to send message to chatID=%d: %v", chatID, err)
			// Можно пометить отдельного получателя как "failed" и т.д.
		} else {
			log.Printf("Сообщение успешно отправлено в чат ID=%d", chatID)
		}
	}

	log.Printf("Задача ID=%s обработана успешно", natsMsg.TaskID)
	return nil
}
