
# **Микросервис рассылки сообщений для Telegram**

[![GitHub](https://img.shields.io/badge/GoBlast-GitHub-blue?logo=github)](https://github.com/0sokrat0/GoBlast)
[![Telegram](https://img.shields.io/badge/GoBlast-Telegram-blue?logo=telegram)](https://t.me/SOKRAT_00)
[![Prometheus](https://img.shields.io/badge/Monitoring-Prometheus-orange?logo=prometheus)](https://prometheus.io/)
[![Grafana](https://img.shields.io/badge/Dashboards-Grafana-green?logo=grafana)](https://grafana.com/)

![GoBlast](https://blog.jetbrains.com/wp-content/uploads/2021/02/Go_8001611039611515.gif)

**GoBlast** — это микросервис для массовой рассылки сообщений через Telegram с использованием **NATS** для очередей и **Redis** для дедупликации. Проект построен на основе **Standard Go Project Layout** и поддерживает многопользовательскую работу с приоритетами и мониторингом.

---

## **Основной функционал**

1. **Поддержка всех типов сообщений**:
    - Текст (text)
    - Фото (photo)
    - Видео (video)
    - Голосовые сообщения (voice)
    - Анимации (animation)
    - Документы (document)

2. **Многопользовательская поддержка**:
    - Авторизация через access_token.
    - Изоляция задач между пользователями.

3. **Регулируемая скорость**:
    - Поддержка приоритетов (high, medium, low) для оптимальной нагрузки.

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

## **Установка и запуск**

### **Системные требования**

- **Go**: >=1.20
- **Docker**: >=20.10
- **Redis**
- **NATS**
- **Prometheus** и **Grafana**

### **Запуск проекта**

1. Клонировать репозиторий:
    ```bash
    git clone https://github.com/0sokrat0/GoBlast.git
    cd GoBlast
    ```

2. Запустить проект через Docker Compose:
    ```bash
    docker-compose up --build
    ```

3. Открыть метрики и дашборды:
    - **Prometheus**: [http://localhost:9090](http://localhost:9090)
    - **Grafana**: [http://localhost:3000](http://localhost:3000)

---

## **API Документация**

Документация доступна в формате Swagger:
- URL: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## **Мониторинг**

В проект включены метрики Prometheus:
- HTTP запросы: `/metrics`
- Мониторинг Telegram-воркеров и задач
- Grafana дашборды для визуализации.

---

## **Контакты**

- Telegram: [@SOKRAT_00](https://t.me/SOKRAT_00)
- GitHub: [0sokrat0](https://github.com/0sokrat0)
