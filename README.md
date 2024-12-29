
---
```plaintext
project/
├── api/                    # HTTP API
│   ├── handlers/           # Обработчики HTTP-запросов
│   │   ├── tasks.go        # Эндпоинты для работы с задачами
│   │   ├── auth.go         # Эндпоинты для авторизации
│   │   ├── status.go       # Эндпоинты для проверки статусов
│   ├── middleware/         # HTTP middleware
│   │   ├── auth.go         # Проверка токена пользователя
│   │   ├── logging.go      # Логирование запросов
│   ├── routes.go           # Роутинг приложения
├── cmd/                    # Основные точки входа
│   ├── api/                # HTTP-сервер API
│   │   └── main.go         # Запуск HTTP-сервера
│   ├── worker/             # Запуск воркеров
│   │   └── main.go         # Запуск обработчиков задач
├── configs/                # Конфигурационные файлы
│   ├── config.yaml         # Основной конфиг
├── internal/               # Внутренняя бизнес-логика
│   ├── tasks/              # Логика работы с задачами
│   │   ├── service.go      # Основные операции с задачами
│   │   ├── deduplication.go# Логика дедупликации
│   │   ├── retry.go        # Логика повторных попыток
│   ├── users/              # Логика работы с пользователями
│   │   ├── service.go      # Логика авторизации и управления пользователями
│   ├── telegram/           # Работа с Telegram API
│   │   ├── client.go       # Telegram клиент
│   │   ├── messages.go     # Отправка сообщений
│   ├── metrics/            # Метрики
│   │   ├── prometheus.go   # Интеграция с Prometheus
├── pkg/                    # Внешние библиотеки (можно переиспользовать)
│   ├── logger/             # Логирование
│   │   ├── logger.go       # Конфигурация logrus
│   ├── queue/              # Работа с NATS
│   │   ├── nats.go         # Подключение к NATS
│   ├── storage/            # Работа с Redis и БД
│   │   ├── redis.go        # Подключение к Redis
│   │   ├── database.go     # Работа с базой данных
├── test/                   # Тесты
│   ├── integration/        # Интеграционные тесты
│   │   ├── tasks_test.go   # Тесты задач
│   ├── unit/               # Юнит-тесты
│   │   ├── tasks_test.go   # Тесты задач
│   │   ├── auth_test.go    # Тесты авторизации
├── Makefile                # Автоматизация сборки
├── go.mod                  # Модуль Go
└── README.md               # Документация
```
---