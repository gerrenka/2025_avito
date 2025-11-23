.PHONY: help build up down restart run rebuild fmt test test-integration

# Переменные
DOCKER_COMPOSE=docker-compose
GO=go

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

# === Docker команды ===

build: ## Собрать Docker образы
	@echo "$(GREEN)Сборка Docker образов...$(NC)"
	$(DOCKER_COMPOSE) build
	@echo "$(GREEN)Сборка завершена$(NC)"

up: ## Запустить все сервисы (PostgreSQL, миграции, приложение)
	@echo "$(GREEN)Запуск сервисов...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Сервисы запущены!$(NC)"
	@echo "API доступен на http://localhost:8080"

down: ## Остановить все сервисы
	@echo "$(YELLOW)Остановка сервисов...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)Сервисы остановлены$(NC)"

restart: down up ## Перезапустить все сервисы

rebuild: down build up ## Пересобрать и перезапустить все сервисы

# === Go команды ===

run: ## Запустить приложение локально (без Docker)
	@echo "$(GREEN)Запуск приложения локально...$(NC)"
	@echo "$(YELLOW)Убедитесь что PostgreSQL запущен!$(NC)"
	$(GO) run cmd/api/main.go

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)Форматирование завершено$(NC)"

test-integration: ## Запустить integration тесты
	@echo "$(GREEN)Запуск integration тестов...$(NC)"
	@echo "$(YELLOW)Убедитесь что PostgreSQL запущен на порту 5433!$(NC)"
	$(GO) test -v -timeout 30s ./test/integration
	@echo "$(GREEN)Integration тесты завершены$(NC)"
