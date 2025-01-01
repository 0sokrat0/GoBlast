package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"GoBlast/configs"
	"GoBlast/internal/tasks"
	"GoBlast/pkg/queue"
	"GoBlast/pkg/storage/db"

	"github.com/nats-io/nats.go"
)

type TaskMessage struct {
	TaskID string `json:"id"`
	// Можно добавить нужные поля (UserID, Content и т.д.)
}

func main() {
	// 1. Загружаем конфиг
	cfg, err := configs.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SslMode,
	)

	// 2. Инициируем БД
	err = db.InitDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.CloseDB()

	// 3. Инициируем NATS
	natsClient, err := queue.NewNatsClient(cfg.Broker.URL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Conn.Close()

	// 4. Подписываемся на "tasks.create"
	_, err = natsClient.Conn.QueueSubscribe("tasks.create", "worker-group", func(msg *nats.Msg) {
		// a) Распарсим сообщение
		var taskMsg TaskMessage
		if err := json.Unmarshal(msg.Data, &taskMsg); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			return
		}

		log.Printf("Получено событие о задаче: %v", taskMsg.TaskID)

		// b) Забираем задачу из БД (например, через tasks.NewTasksRepository)
		repository := tasks.NewTasksRepository(db.DB)
		currentTask, err := repository.GetTaskByID(taskMsg.TaskID)
		if err != nil {
			log.Printf("Failed to find task %s: %v", taskMsg.TaskID, err)
			return
		}

		// c) Обрабатываем задачу (отправляем в Telegram)
		err = processTask(currentTask)
		if err != nil {
			log.Printf("Task processing error: %v", err)
			// Можно обновить статус задачи в БД (например, "failed")
			return
		}

		// d) При успехе — обновляем статус задачи в БД
		currentTask.Status = "completed"
		repository.SaveTask(currentTask)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to tasks.create: %v", err)
	}

	log.Println("Worker is running...")

	// 5. Ждём сигнал остановки (Ctrl+C, kill)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Worker stopped.")
}

func processTask() error {
	// Здесь логика отправки в Telegram
	// (Можно использовать внутренние пакеты, например, internal/telegram/client.go)
	// Пример (очень упрощённо):
	// err := telegram.SendMessage(t.Recipients, t.Content, ...)
	// return err
	return nil
}
