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

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const batchSize = 100

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

// SubscribeTasks - подписка на задачи из NATS// SubscribeTasks - подписка на задачи из NATS
// internal/worker/worker.go

func SubscribeTasks(natsClient *queue.NATSClient, db *gorm.DB, botManager *BotManager) error {
	_, err := natsClient.Conn.QueueSubscribe("tasks.create", "worker-group", func(msg *nats.Msg) {
		var natsMsg TaskNATSMessage
		if err := json.Unmarshal(msg.Data, &natsMsg); err != nil {
			logger.Log.Error("Ошибка десериализации сообщения NATS", zap.Error(err))
			return
		}

		// Проверка корректности данных
		if err := validateTaskMessage(natsMsg); err != nil {
			logger.Log.Error("Некорректное сообщение NATS", zap.Error(err))
			return
		}

		// Получаем токен бота
		userRepo := users.NewAuthUserRepository(db)
		userData, err := userRepo.FindByID(natsMsg.UserID)
		if err != nil {
			logger.Log.Error("Ошибка получения пользователя", zap.Error(err), zap.Uint("user_id", natsMsg.UserID))
			return
		}

		botToken, err := decryptToken(userData.Token)
		if err != nil {
			logger.Log.Error("Ошибка дешифрования токена", zap.Error(err), zap.Uint("user_id", natsMsg.UserID))
			return
		}

		// Передаем задачу в BotManager
		botManager.StartTask(botToken, natsMsg, db)
	})

	if err != nil {
		return err
	}

	logger.Log.Info("Подписка на NATS успешно выполнена")
	return nil
}

func validateTaskMessage(task TaskNATSMessage) error {
	if task.TaskID == "" {
		return errors.New("пустой TaskID")
	}
	if task.UserID == 0 {
		return errors.New("пустой UserID")
	}
	if len(task.Recipients) == 0 {
		return errors.New("пустой список получателей")
	}
	if task.Content.Type == "" {
		return errors.New("пустой тип контента")
	}
	return nil
}

func decryptToken(encryptedToken string) (string, error) {
	// Декодируем base64
	decodedToken, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}

	// Дешифруем токен
	decryptedTokenBytes, err := encryption.Decrypt(decodedToken, []byte(middleware.EncryptionKey))
	if err != nil {
		return "", err
	}

	return string(decryptedTokenBytes), nil
}
