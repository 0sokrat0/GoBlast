# **Микросервис рассылки сообщений для Telegram**

[![GitHub](https://img.shields.io/badge/GoBlast-GitHub-blue?logo=github)](https://github.com/0sokrat0/GoBlast)
[![Telegram](https://img.shields.io/badge/GoBlast-Telegram-blue?logo=telegram)](https://t.me/SOKRAT_00)



![GoBlast](https://blog.jetbrains.com/wp-content/uploads/2021/02/Go_8001611039611515.gif)

**GoBlast** — это микросервис для массовой рассылки сообщений через Telegram с использованием **NATS** для очередей и **Redis** для дедупликации. Проект построен на основе **Standard Go Project Layout** и поддерживает многопользовательскую работу с приоритетами и мониторингом.

---

## **Основной функционал**

1. **Поддержка всех типов сообщений**:
    - Текст (`text`)
    - Фото (`photo`)
    - Видео (`video`)
    - Голосовые сообщения (`voice`)
    - Анимации (`animation`)
    - Документы (`document`)

2. **Многопользовательская поддержка**:
    - Авторизация через `access_token`.
    - Изоляция задач между пользователями.

3. **Регулируемая скорость**:
    - Поддержка приоритетов (`high`, `medium`, `low`) для оптимальной нагрузки.

4. **Дедупликация сообщений**:
    - Проверка уникальности сообщений с использованием Redis.

5. **Асинхронная обработка**:
    - Очереди задач через **NATS**.
    - Параллельная обработка задач с помощью воркеров.

6. **Мониторинг и логирование**:
    - Метрики **Prometheus**: количество задач, время обработки, ошибки.
    - Логирование через **Zap**.

---

## **Архитектура проекта**

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
       | Telegram API   | |  Redis / DB     | |   Логирование  |
       +----------------+ +----------------+ +----------------+
       
```
---
```
.
├── cmd                     # Точка входа приложения
│   └── main.go
├── configs                 # Конфигурационные файлы
│   ├── config.go
│   └── config.yaml
├── docker-compose.yaml     # Настройки Docker Compose
├── Dockerfile              # Docker-образ
├── docs                    # Документация Swagger
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal                # Внутренняя логика микросервиса
│   ├── api
│   │   ├── handlers        # Обработчики маршрутов
│   │   ├── middleware      # Middleware (аутентификация, логирование)
│   │   └── routes.go       # Регистрация маршрутов
│   ├── metrics             # Метрики Prometheus
│   ├── tasks               # Бизнес-логика задач
│   ├── telegram            # Работа с Telegram API
│   ├── users               # Логика аутентификации
│   └── worker              # Воркеры для обработки задач
├── pkg                     # Общие библиотеки
│   ├── encryption          # Шифрование
│   ├── logger              # Логирование
│   ├── queue               # Очереди NATS
│   ├── response            # Форматирование API ответов
│   └── storage             # Работа с базой данных
├── README.md               # Документация проекта
└── test                    # Тесты
    ├── integration         # Интеграционные тесты
    └── unit                # Юнит-тесты
```
---

Эндпоинты API
1. Регистрация пользователя

Регистрация нового пользователя.

Метод: POST /register

    Тело запроса:

{
  "username": "example_user",
  "password": "password123"
}

Ответ:

    {
      "success": true,
      "data": {
        "user_id": 1
      }
    }
---
2. Авторизация

Получение токена access_token.

Метод: POST /login

    Тело запроса:

{
  "username": "example_user",
  "password": "password123"
}

Ответ:

    {
      "success": true,
      "data": {
        "access_token": "eyJhbGciOiJIUzI1..."
      }
    }
---
3. Создание задачи

Создание задачи для массовой рассылки.

Метод: POST /bulk_message

    Тело запроса:

{
  "recipients": [123456789, 987654321],
  "content": {
    "type": "text",
    "text": "Привет, это тестовое сообщение!"
  },
  "priority": "high"
}

Ответ:

    {
      "success": true,
      "data": {
        "task_id": "a804bd98-8e4d-4e8d-9678-7e28b7a8408f",
        "status": "scheduled"
      }
    }
---
4. Проверка статуса задачи

Получение информации о задаче.

Метод: GET /status/{task_id}

    Пример ответа:

    {
      "success": true,
      "data": {
        "task_id": "a804bd98-8e4d-4e8d-9678-7e28b7a8408f",
        "status": "completed",
        "recipients_count": 2
      }
    }
---
5. Отмена задачи

Отмена задачи по ID.

Метод: POST /cancel/{task_id}

    Пример ответа:

    {
      "success": true,
      "data": {
        "task_id": "a804bd98-8e4d-4e8d-9678-7e28b7a8408f",
        "status": "cancelled"
      }
    }

Мониторинг

    Метрики Prometheus: GET /metrics
    Swagger UI: GET /swagger/index.html

