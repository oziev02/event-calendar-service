# Event Calendar Service

HTTP-сервер для работы с календарем событий. Реализован с использованием чистой архитектуры, DDD принципов, SOLID, KISS, DRY.

## Архитектура

Проект использует стандартную Go архитектуру, принятую в big tech компаниях:

- **model** (`internal/model/`) - доменные модели, интерфейсы, ошибки
- **service** (`internal/service/`) - бизнес-логика
- **repository** (`internal/repository/`) - слой хранения данных
- **handler** (`internal/handler/`) - HTTP обработчики
- **middleware** (`internal/middleware/`) - HTTP middleware
- **worker** (`internal/worker/`) - фоновые воркеры
- **config** (`internal/config/`) - конфигурация
- **logger** (`internal/logger/`) - логирование

## Особенности

- CRUD операции для событий
- Фоновый воркер для напоминаний через канал
- Автоматическая архивация старых событий
- Асинхронное логирование через канал
- Middleware для логирования HTTP запросов
- Unit тесты для бизнес-логики
- Docker поддержка

## Требования

- Go 1.21+
- Docker и Docker Compose (опционально)

## Установка и запуск

### Локальный запуск

```bash
# Установка зависимостей
go mod download

# Запуск сервера
go run cmd/server/main.go

# Или с флагами
go run cmd/server/main.go -port=8080
```

### Запуск через Docker

```bash
# Сборка и запуск
docker-compose up --build

# В фоновом режиме
docker-compose up -d --build
```

### Конфигурация

Конфигурация загружается из `.env` файла. Скопируйте `.env.example` в `.env` и настройте параметры:

```bash
cp .env.example .env
```

Переменные окружения в `.env`:

- `PORT` - порт сервера (по умолчанию: 8080)
- `CLEANUP_INTERVAL` - интервал очистки старых событий (по умолчанию: 5m)
- `ARCHIVE_AFTER` - время до архивации события (по умолчанию: 720h = 30 дней)
- `REMINDER_CHECK_INTERVAL` - интервал проверки напоминаний (по умолчанию: 1m)
- `LOGGER_BUFFER_SIZE` - размер буфера логгера (по умолчанию: 100)

Также можно переопределить значения через переменные окружения системы или флаги командной строки.

## API Endpoints

### POST /create_event

Создание нового события.

**Формат запроса (JSON):**
```json
{
  "user_id": "user1",
  "date": "2024-01-15",
  "event": "Встреча с командой",
  "reminder_time": "2024-01-15T09:00:00Z"
}
```

**Формат запроса (Form Data):**
```
user_id=user1&date=2024-01-15&event=Встреча с командой&reminder_time=2024-01-15T09:00:00Z
```

**Пример запроса:**
```bash
# JSON
curl -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "date": "2024-01-15",
    "event": "Встреча с командой",
    "reminder_time": "2024-01-15T09:00:00Z"
  }'

# Form Data
curl -X POST http://localhost:8080/create_event \
  -d "user_id=user1&date=2024-01-15&event=Встреча с командой"
```

**Ответ (успех):**
```json
{
  "result": {
    "event_id": "20240115120000-abc123",
    "message": "Event created successfully"
  }
}
```

**Ответ (ошибка):**
```json
{
  "error": "Invalid date format. Use YYYY-MM-DD"
}
```

### POST /update_event

Обновление существующего события.

**Формат запроса (JSON):**
```json
{
  "user_id": "user1",
  "event_id": "20240115120000-abc123",
  "date": "2024-01-16",
  "event": "Обновленная встреча",
  "reminder_time": "2024-01-16T09:00:00Z"
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/update_event \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "event_id": "20240115120000-abc123",
    "date": "2024-01-16",
    "event": "Обновленная встреча"
  }'
```

### POST /delete_event

Удаление события.

**Формат запроса (JSON):**
```json
{
  "user_id": "user1",
  "event_id": "20240115120000-abc123"
}
```

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/delete_event \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user1",
    "event_id": "20240115120000-abc123"
  }'
```

### GET /events_for_day

Получение всех событий на день.

**Параметры запроса:**
- `user_id` - идентификатор пользователя (обязательно)
- `date` - дата в формате YYYY-MM-DD (обязательно)

**Пример запроса:**
```bash
curl "http://localhost:8080/events_for_day?user_id=user1&date=2024-01-15"
```

**Ответ:**
```json
{
  "result": {
    "events": [
      {
        "id": "20240115120000-abc123",
        "user_id": "user1",
        "date": "2024-01-15",
        "text": "Встреча с командой",
        "reminder_time": "2024-01-15T09:00:00Z",
        "created_at": "2024-01-15T12:00:00Z",
        "updated_at": "2024-01-15T12:00:00Z"
      }
    ]
  }
}
```

### GET /events_for_week

Получение событий на неделю (начиная с указанной даты).

**Параметры запроса:**
- `user_id` - идентификатор пользователя (обязательно)
- `date` - начальная дата недели в формате YYYY-MM-DD (обязательно)

**Пример запроса:**
```bash
curl "http://localhost:8080/events_for_week?user_id=user1&date=2024-01-15"
```

### GET /events_for_month

Получение событий на месяц.

**Параметры запроса:**
- `user_id` - идентификатор пользователя (обязательно)
- `date` - любая дата месяца в формате YYYY-MM-DD (обязательно)

**Пример запроса:**
```bash
curl "http://localhost:8080/events_for_month?user_id=user1&date=2024-01-15"
```

## HTTP Status Codes

- `200 OK` - успешный запрос
- `400 Bad Request` - ошибка валидации (некорректный формат даты, отсутствующие поля)
- `503 Service Unavailable` - ошибка бизнес-логики (событие не найдено)
- `500 Internal Server Error` - внутренняя ошибка сервера

## Фоновые воркеры

### Reminder Worker

Воркер обрабатывает напоминания о событиях через канал. При создании события с `reminder_time`, задача добавляется в канал, и воркер отслеживает время и отправляет напоминания.

### Cleanup Worker

Отдельная горутина, которая каждые X минут (настраивается через `CLEANUP_INTERVAL`) архивирует старые события (старше `ARCHIVE_AFTER`).

### Async Logger

HTTP handlers не пишут в stdout напрямую, а отправляют записи в канал, который обрабатывает отдельная горутина для асинхронного логирования.

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с покрытием
go test -cover ./...

# Запуск тестов конкретного пакета
go test ./internal/application/...

# Запуск тестов с verbose
go test -v ./...
```

## Проверка кода

```bash
# Форматирование кода
go fmt ./...

# Статический анализ
go vet ./...

# Линтинг (если установлен golangci-lint)
golangci-lint run
```

## Структура проекта

```
event-calendar-service/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── model/                    # Доменные модели
│   │   ├── event.go
│   │   └── reminder.go
│   ├── service/                  # Бизнес-логика
│   │   ├── event_service.go
│   │   └── event_service_test.go
│   ├── repository/              # Слой хранения данных
│   │   └── memory_repository.go
│   ├── handler/                 # HTTP обработчики
│   │   ├── event_handler.go
│   │   └── form_decoder.go
│   ├── middleware/              # HTTP middleware
│   │   └── logging_middleware.go
│   ├── worker/                  # Фоновые воркеры
│   │   ├── reminder_worker.go
│   │   └── cleanup_worker.go
│   ├── reminder/                # Отправка напоминаний
│   │   └── sender.go
│   ├── config/                  # Конфигурация
│   │   └── config.go
│   ├── logger/                  # Логирование
│   │   └── logger.go
│   └── server.go                # HTTP сервер
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Принципы проектирования

- **Clean Architecture** - разделение на слои с четкими зависимостями
- **DDD** - доменно-ориентированное проектирование
- **SOLID** - следование принципам SOLID
- **KISS** - простота решения
- **DRY** - избегание дублирования кода
- **Error Handling** - использование `errors.Is` и `errors.As` для сравнения ошибок

## Лицензия

MIT

