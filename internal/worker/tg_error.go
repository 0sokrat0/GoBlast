package worker

import (
	"GoBlast/pkg/logger"
	"fmt"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrorMapping — карта действий для различных типов ошибок.
// Допустим, если "chat not found (400)", мы распознаём как "NOT_FOUND".
var ErrorMapping = map[string]func(*Worker, TaskItem, error){
	"FLOOD_WAIT":     handleFloodWait,
	"UNAUTHORIZED":   handleUnauthorized,
	"NOT_FOUND":      handleNotFound, // если хотим
	"BAD_REQUEST":    handleBadRequest,
	"INTERNAL_ERROR": handleInternalError,
	"default":        handleDefaultError,
}

// handleTgError разбирает err.Error(), ищет ключевые слова, вызывает соответствующий хендлер.
func (w *Worker) handleTgError(item TaskItem, err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	logger.Log.Warn("[Worker] Обнаружена ошибка Telegram API",
		zap.String("task_id", item.TaskID),
		zap.Int64("recipient", item.Recipient),
		zap.String("error", msg))

	// Если видим «chat not found (400)», считаем это "NOT_FOUND"
	if strings.Contains(msg, "chat not found (400)") {
		handleNotFound(w, item, err)
		return err
	}
	// Аналогично, если хотим анализировать "400 Bad Request" и т. п.

	// Иначе ищем по ключам ErrorMapping
	for key, handler := range ErrorMapping {
		if strings.Contains(msg, key) {
			handler(w, item, err)
			return err
		}
	}

	// Иначе default
	ErrorMapping["default"](w, item, err)
	return err
}

func parseFloodWait(msg string) int {
	re := regexp.MustCompile(`FLOOD_WAIT_(\d+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) == 2 {
		sec, _ := strconv.Atoi(matches[1])
		return sec
	}
	return 0
}
func handleFloodWait(w *Worker, item TaskItem, err error) {
	waitSeconds := parseFloodWait(err.Error())
	if waitSeconds > 0 {
		logger.Log.Warn("[Worker] FLOOD WAIT обнаружен, ожидаем",
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient),
			zap.Int("wait_seconds", waitSeconds))

		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}
	// По окончании всё равно incrementFailed
	w.incrementFailed(item.TaskID, item.Content.Type, err)
}

// handleUnauthorized
func handleUnauthorized(w *Worker, item TaskItem, err error) {
	logger.Log.Error("[Worker] UNAUTHORIZED ошибка, завершение обработки получателя",
		zap.String("task_id", item.TaskID),
		zap.Int64("recipient", item.Recipient),
		zap.Error(err))

	notifyAdmin(fmt.Sprintf("UNAUTHORIZED для задачи %s, получатель %d",
		item.TaskID, item.Recipient))

	w.incrementFailed(item.TaskID, item.Content.Type, err)
}

// handleNotFound — как пример
func handleNotFound(w *Worker, item TaskItem, err error) {
	logger.Log.Warn("[Worker] NOT_FOUND (chat not found)",
		zap.String("task_id", item.TaskID),
		zap.Int64("recipient", item.Recipient),
		zap.Error(err))
	w.incrementFailed(item.TaskID, item.Content.Type, err)
}

// handleBadRequest
func handleBadRequest(w *Worker, item TaskItem, err error) {
	logger.Log.Warn("[Worker] BAD_REQUEST",
		zap.String("task_id", item.TaskID),
		zap.Int64("recipient", item.Recipient),
		zap.Error(err))
	w.incrementFailed(item.TaskID, item.Content.Type, err)
}

// handleInternalError
func handleInternalError(w *Worker, item TaskItem, err error) {
	logger.Log.Error("[Worker] INTERNAL_ERROR (Telegram)",
		zap.String("task_id", item.TaskID),
		zap.Int64("recipient", item.Recipient),
		zap.Error(err))

	// повторно отправим через 5 секунд
	go func() {
		time.Sleep(5 * time.Second)
		w.TaskChan <- item
	}()
}

// handleDefaultError
func handleDefaultError(w *Worker, item TaskItem, err error) {
	logger.Log.Error("[Worker] Неизвестная ошибка Telegram",
		zap.String("task_id", item.TaskID),
		zap.Int64("recipient", item.Recipient),
		zap.Error(err))

	w.incrementFailed(item.TaskID, item.Content.Type, err)
}

// notifyAdmin — отправить уведомление
func notifyAdmin(message string) {
	logger.Log.Warn("[AdminNotify] Уведомление администратора",
		zap.String("message", message))
}
