.PHONY: up up-bg down restart logs

# Запуск сервисов
up:
	docker-compose up --build

# Запуск в фоновом режиме
up-bg:
	docker-compose up --build -d

# Остановка сервисов
down:
	docker-compose down

# Пересборка и перезапуск сервисов
restart: down up

# Просмотр логов
logs:
	docker-compose logs -f