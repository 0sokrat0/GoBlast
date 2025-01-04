package main

import (
	"GoBlast/configs"
	"GoBlast/internal/api"
	"GoBlast/internal/api/middleware"
	"GoBlast/internal/worker"
	"GoBlast/pkg/logger"
	"GoBlast/pkg/queue"
	"GoBlast/pkg/storage/db"
	"context"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"

	_ "GoBlast/docs"
)

// @title GoBlast API
// @version 1.0
// @description API для управления рассылками с поддержкой JWT аутентификации.
// @termsOfService http://example.com/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "Введите JWT токен в формате: Bearer {your-token}"
func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "production"
	}
	if err := logger.InitLogger(env); err != nil {
		fmt.Printf("Не удалось инициализировать логгер: %v\n", err)
		os.Exit(1)
	}
	defer logger.SyncLogger()

	// Обработка сигналов
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Инициализация конфигурации
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs"
	}
	cfg, err := configs.LoadConfig(configPath)
	if err != nil {
		logger.Log.Fatal("Ошибка загрузки конфигурации", zap.Error(err))
	}

	logger.Log.Info("Приложение запущено",
		zap.String("environment", cfg.App.Environment),
		zap.Int("port", cfg.App.Port),
	)

	// Инициализация базы данных
	dbConn := initDB(cfg)
	defer func() {
		sqlDB, err := dbConn.DB()
		if err != nil {
			logger.Log.Error("Ошибка получения sql.DB", zap.Error(err))
			return
		}
		if err := sqlDB.Close(); err != nil {
			logger.Log.Error("Ошибка закрытия соединения с БД", zap.Error(err))
		}
	}()

	// Инициализация NATS
	natsClient := initQueue(cfg)
	defer func() {
		natsClient.Conn.Close()
		logger.Log.Info("Соединение с NATS закрыто")
	}()

	// Инициализация middleware
	middleware.Initialize(cfg)

	// Запуск сервера и воркера
	go startServer(ctx, cfg, dbConn, natsClient)
	startWorker(ctx, cfg, dbConn, natsClient)

	logger.Log.Info("Программа завершена")
}

func initDB(cfg *configs.Config) *gorm.DB {
	dsn := configs.GetDSN(cfg.Database)
	err := db.InitDB(dsn)
	if err != nil {
		logger.Log.Fatal("Ошибка инициализации БД", zap.Error(err)) // Завершаем выполнение
	}
	logger.Log.Info("Соединение с БД установлено", zap.String("dsn", dsn))
	return db.DB
}

func initQueue(cfg *configs.Config) *queue.NATSClient {
	client, err := queue.NewNatsClient(cfg.Broker.URL)
	if err != nil {
		logger.Log.Fatal("Ошибка подключения к NATS", zap.Error(err)) // Завершаем выполнение
	}
	logger.Log.Info("Соединение с NATS установлено", zap.String("url", cfg.Broker.URL))
	return client
}

func startServer(ctx context.Context, cfg *configs.Config, db *gorm.DB, natsClient *queue.NATSClient) {
	router := api.SetupRouter(db, natsClient)
	server := fmt.Sprintf(":%d", cfg.App.Port)

	go func() {
		if err := router.Run(server); err != nil {
			logger.Log.Error("Ошибка запуска сервера", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Log.Info("Сервер завершает работу...")
}

func startWorker(ctx context.Context, cfg *configs.Config, db *gorm.DB, natsClient *queue.NATSClient) {
	go func() {
		logger.Log.Info("Воркер запущен")
		if err := worker.Worker(cfg, db, natsClient); err != nil {
			logger.Log.Error("Ошибка в воркере", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Log.Info("Воркер завершает работу...")
}
