package worker

import (
	"GoBlast/internal/api/middleware"
	"GoBlast/internal/users"
	"GoBlast/pkg/encryption"
	"GoBlast/pkg/logger"
	"GoBlast/pkg/queue"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TaskNATSMessage — структура задачи, приходящей из NATS.
type TaskNATSMessage struct {
	TaskID     string  `json:"task_id"` // ID задачи
	UserID     uint    `json:"user_id"`
	Recipients []int64 `json:"recipients"`
	Content    Content `json:"content"`
	Priority   string  `json:"priority,omitempty"`
	// Schedule ... (если нужно)
}

// Content — описание контента (тип, текст/медиа и т. д.)
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	MediaURL string `json:"media_url"`
	MediaID  string `json:"media_id"`
	Caption  string `json:"caption"`
}

func SubscribeTasks(natsClient *queue.NATSClient, db *gorm.DB, botManager *BotManager) error {
	_, err := natsClient.Conn.QueueSubscribe("tasks.create", "worker-group", func(msg *nats.Msg) {
		var natsMsg TaskNATSMessage
		if e := json.Unmarshal(msg.Data, &natsMsg); e != nil {
			logger.Log.Error("Ошибка десериализации сообщения NATS", zap.Error(e))
			return
		}
		logger.Log.Info("[Subscriber] Получено сообщение NATS",
			zap.String("task_id", natsMsg.TaskID),
			zap.Uint("user_id", natsMsg.UserID),
			zap.Int("recipients_count", len(natsMsg.Recipients)),
			zap.String("priority", natsMsg.Priority))

		// Валидация
		if e := validateTaskMessage(natsMsg); e != nil {
			logger.Log.Error("Некорректное сообщение NATS", zap.Error(e))
			return
		}

		// Ищем в БД токен бота
		userRepo := users.NewAuthUserRepository(db)
		userData, e := userRepo.FindByID(natsMsg.UserID)
		if e != nil {
			logger.Log.Error("Ошибка получения пользователя",
				zap.Error(e),
				zap.Uint("user_id", natsMsg.UserID))
			return
		}

		botToken, e := decryptToken(userData.Token)
		if e != nil {
			logger.Log.Error("Ошибка дешифрования токена",
				zap.Error(e),
				zap.Uint("user_id", natsMsg.UserID))
			return
		}

		// Передаём задачу в BotManager
		botManager.StartTask(botToken, natsMsg, db)
	})
	if err != nil {
		return err
	}

	logger.Log.Info("Подписка на NATS успешно выполнена")
	return nil
}

// validateTaskMessage проверяет ключевые поля
func validateTaskMessage(task TaskNATSMessage) error {
	if task.TaskID == "" {
		return errors.New("пустой task_id")
	}
	if task.UserID == 0 {
		return errors.New("пустой user_id")
	}
	if len(task.Recipients) == 0 {
		return errors.New("пустой список получателей")
	}
	if strings.TrimSpace(task.Content.Type) == "" {
		return errors.New("пустой тип контента")
	}
	return nil
}

// decryptToken расшифровывает токен бота из userData.Token
func decryptToken(encryptedToken string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}
	plain, err := encryption.Decrypt(decoded, []byte(middleware.EncryptionKey))
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
