# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы.

## Быстрый старт

### Запуск сервиса

```bash
make up
```

Сервис будет доступен на `http://localhost:8080`

### Остановка сервиса

```bash
make down
```

## Доступные команды (Makefile)

| Команда | Описание |
|---------|----------|
| `make build` | Собрать Docker образы |
| `make up` | Запустить все сервисы |
| `make down` | Остановить все сервисы |
| `make restart` | Перезапустить все сервисы |
| `make rebuild` | Пересобрать и перезапустить |
| `make run` | Запустить локально (без Docker) |
| `make fmt` | Форматировать код |
| `make test-integration` | Запустить integration тесты |

## Локальная разработка

```bash
make run
```

Требуется запущенный PostgreSQL на порту 5432.

## Тестирование

```bash
make test-integration
```

Требуется запущенный PostgreSQL на порту 5433.

## Стек

- Go 1.21
- PostgreSQL 15
- Docker / Docker Compose

## Принятые решения

### 1. Получение ревью для пользователя без назначений
Эндпоинт `/users/getReview` возвращает 404 если пользователь не существует в системе. Для существующего пользователя без назначений возвращается пустой массив `pull_requests: []`.

### 2. Эндпоинт `/health`
В OpenAPI указан тег Health без описания конкретного пути. Реализован GET `/health` возвращающий `{"status":"ok"}`.




### 🏗️ Архитектура

```
review-service/
├── cmd/api/                    # Точка входа приложения
│   └── main.go
├── internal/
│   ├── config/                 # Конфигурация
│   │   └── config.go
│   ├── domain/                 # Бизнес-сущности
│   │   ├── user.go
│   │   ├── team.go
│   │   ├── pull_request.go
│   │   └── errors.go
│   ├── handlers/               # HTTP обработчики
│   │   ├── team.go
│   │   ├── user.go
│   │   ├── pull_request.go
│   │   ├── router.go
│   │   ├── health.go
│   │   ├── helpers.go
│   │   └── middleware/
│   │       ├── logger.go
│   │       └── recovery.go
│   ├── repository/             # Слой данных
│   │   ├── interfaces.go
│   │   └── postgres/
│   │       ├── user_repository.go
│   │       ├── team_repository.go
│   │       └── pr_repository.go
│   └── service/                # Бизнес-логика
│       ├── service_interfaces.go
│       ├── user_service.go
│       ├── team_service.go
│       └── pr_service.go
├── migrations/                 # Миграции БД
│   ├── 001_init_schema.up.sql
│   └── 001_init_schema.down.sql
├── test/integration/
│   └── integration_test.go
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── env.example
├── go.mod
├── go.sum
└── README.md
```