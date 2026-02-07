# Rural Portal

Go API приложение для сельского портала с чистой архитектурой.


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
