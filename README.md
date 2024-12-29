# Микросервис рассылки сообщений для Telegram

![GoBlast](https://blog.jetbrains.com/wp-content/uploads/2021/02/Go_8001611039611515.gif)

Этот микросервис предназначен для массовой рассылки сообщений через Telegram с использованием NATS для очередей и Redis для дедупликации. Проект построен на основе **Standard Go Project Layout**
---

## Основной функционал

1. **Поддержка всех типов сообщений**:
   - Текст (`text`)
   - Фото (`photo`)
   - Видео (`video`)
   - Голосовые сообщения (`voice`)
   - Кружки (`animation`)
   - Документы (`document`)

2. **Многопользовательская поддержка**:
   - Авторизация через `access_token`.
   - Изоляция задач по пользователям.

3. **Регулируемая скорость**:
   - Приоритеты (`high`, `medium`, `low`) для оптимизации нагрузки.

4. **Дедупликация**:
   - Проверка уникальности сообщений через Redis.

5. **Асинхронная обработка**:
   - Очереди задач через NATS.
   - Параллельная обработка с помощью воркеров.

6. **Мониторинг и логи**:
   - Метрики Prometheus (количество задач, время выполнения, ошибки).
   - Логирование через logrus.

---

## Архитектура проекта

```plaintext
+----------------------+      +---------------------------+
|      Клиент          | ---> |          API              |
|  (авторизация, задачи)|      |   (валидация, дедупликация)|
+----------------------+      +---------------------------+
                                    |
                                    v
                        +-----------------------+
                        |         NATS          |
                        +-----------------------+
                                    |
                                    v
                  +---------------------------+
                  |         Workers           |
                  |  (дедупликация, Telegram) |
                  +---------------------------+
                   /             |             \
       +----------------+ +----------------+ +----------------+
       | Telegram API   | |  Redis / DB     | |   Логи         |
       +----------------+ +----------------+ +----------------+
       
```

---

## Структура проекта

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

## Эндпоинты API

1. Регистрация пользователя

POST /register
Регистрация нового пользователя.
2. Авторизация

```
POST /login
```

Получение access_token.

3. Создание задачи

```
POST /bulk_message
```

Создание задачи для рассылки.
4. Проверка статуса задачи

```
GET /status/{task_id}
```

Проверка текущего статуса задачи.
5. Отмена задачи

```
POST /cancel/{task_id}
```

Отмена задачи по ID.

---
