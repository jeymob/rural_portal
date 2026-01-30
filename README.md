# Rural Portal

Go API приложение для сельского портала с чистой архитектурой.

## Структура проекта

```
rural-portal/
├── cmd/
│   └── api/
│       └── main.go              # Точка входа
├── internal/
│   ├── application/
│   │   └── auth/
│   │       └── service.go       # Сервис аутентификации
│   ├── config/
│   │   └── config.go            # Конфигурация
│   ├── delivery/
│   │   └── http/
│   │       ├── handlers/
│   │       │   └── health.go    # HTTP обработчики
│   │       └── routes.go        # Маршруты
│   ├── domain/
│   │   └── user.go              # Доменные модели
│   └── infrastructure/
│       └── persistence/
│           └── models/
│               └── user.go      # Модели БД
├── .env.example                 # Пример переменных окружения
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Установка

1. Клонировать репозиторий
2. Скопировать `.env.example` в `.env`:
```bash
cp .env.example .env
```

3. Установить зависимости:
```bash
go mod download
```

## Запуск

### Локально

```bash
go run ./cmd/api
```

### С Docker

```bash
docker-compose up
```

API будет доступен на `http://localhost:8080`

## Endpoints

- `GET /health` - Проверка статуса приложения

## Разработка

```bash
# Запустить тесты
go test ./...

# Сборка
go build -o app ./cmd/api
```
