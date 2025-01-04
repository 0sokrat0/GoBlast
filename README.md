# Микросервис рассылки сообщений для Telegram
---

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
.
├── cmd
│   └── main.go
├── configs
│   ├── config.go
│   └── config.yaml
├── docker-compose.yaml
├── Dockerfile
├── docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── internal
│   ├── api
│   │   ├── handlers
│   │   │   ├── AuthLogin.go
│   │   │   ├── AuthRegister.go
│   │   │   ├── status.go
│   │   │   └── tasks.go
│   │   ├── middleware
│   │   │   ├── auth.go
│   │   │   ├── CORSM.go
│   │   │   └── logging.go
│   │   └── routes.go
│   ├── metrics
│   ├── routes
│   │   ├── auth.go
│   │   └── tasks.go
│   ├── tasks
│   │   └── tasks.go
│   ├── telegram
│   ├── users
│   │   └── auth.go
│   └── worker
│       └── worker.go
├── Makefile
├── pkg
│   ├── encryption
│   │   ├── encryption.go
│   │   └── encryption_test.go
│   ├── logger
│   │   └── Zap.go
│   ├── queue
│   │   └── nats.go
│   ├── response
│   │   └── apiresponse.go
│   └── storage
│       ├── db
│       │   └── conn.go
│       └── models
│           ├── AuthUser.go
│           └── Tasks.go
├── README.md
└── test
    ├── integration
    └── unit

25 directories, 33 files

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
