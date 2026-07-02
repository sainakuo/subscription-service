# Subscription Service

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Возможности

- CRUDL-операции над подписками
- Подсчёт суммарной стоимости подписок за выбранный период
- Фильтрация по `user_id` и `service_name`
- PostgreSQL для хранения данных
- Миграции для инициализации базы данных
- Конфигурация через `.env`
- Структурированные JSON-логи
- Swagger-документация
- Запуск через Docker Compose

## Стек

- Go
- Gin
- PostgreSQL
- pgx
- golang-migrate
- Docker Compose
- Swagger / Swaggo
- slog

## Структура проекта

```text
.
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── config/
│   ├── db/
│   ├── handler/
│   ├── logger/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   └── service/
├── migrations/
├── docs/
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
└── README.md
```

## Переменные окружения

Пример находится в файле `.env.example`.

```env
APP_PORT=8080
LOG_LEVEL=info

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=subscriptions
DB_SSLMODE=disable
```

Для локального запуска через `go run` используется:

```env
DB_HOST=localhost
```

При запуске через Docker Compose приложение использует:

```env
DB_HOST=postgres
```

Это значение задаётся внутри `docker-compose.yml`.

## Запуск через Docker Compose

Создайте `.env` на основе `.env.example`.

Для Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

Запустите проект:

```powershell
docker compose up --build
```

После запуска будут доступны:

```text
API:     http://localhost:8080
Swagger: http://localhost:8080/swagger/index.html
```

Остановить контейнеры:

```powershell
docker compose down
```

Остановить контейнеры и удалить данные PostgreSQL:

```powershell
docker compose down -v
```

## Локальный запуск без Docker приложения

Можно поднять PostgreSQL через Docker Compose:

```powershell
docker compose up -d postgres
```

Применить миграции:

```powershell
docker compose run --rm migrate
```

Запустить приложение локально:

```powershell
go run ./cmd/app
```

## Swagger

Swagger UI доступен по адресу:

```text
http://localhost:8080/swagger/index.html
```

Если Swagger-документацию нужно пересгенерировать:

```powershell
swag init -g cmd/app/main.go
```

Или, если `swag` не найден:

```powershell
$env:USERPROFILE\go\bin\swag.exe init -g cmd/app/main.go
```

## API

### Health check

```http
GET /health
```

### Создать подписку

```http
POST /subscriptions
```

Пример тела запроса:

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}
```

Пример с датой окончания:

```json
{
  "service_name": "Netflix",
  "price": 799,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "08-2025",
  "end_date": "12-2025"
}
```

### Получить список подписок

```http
GET /subscriptions
```

Доступные query-параметры:

```text
user_id
service_name
limit
offset
```

Примеры:

```http
GET /subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba
GET /subscriptions?service_name=Yandex
GET /subscriptions?limit=10&offset=0
```

### Получить подписку по ID

```http
GET /subscriptions/{id}
```

### Обновить подписку

```http
PUT /subscriptions/{id}
```

Пример тела запроса:

```json
{
  "service_name": "Netflix",
  "price": 799,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "08-2025",
  "end_date": "12-2025"
}
```

### Удалить подписку

```http
DELETE /subscriptions/{id}
```

При успешном удалении возвращается статус:

```text
204 No Content
```

### Подсчитать суммарную стоимость подписок

```http
GET /subscriptions/total-cost
```

Обязательные query-параметры:

```text
from
to
```

Опциональные query-параметры:

```text
user_id
service_name
```

Пример:

```http
GET /subscriptions/total-cost?from=07-2025&to=09-2025
```

Пример с фильтрами:

```http
GET /subscriptions/total-cost?from=07-2025&to=09-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex
```

Пример ответа:

```json
{
  "total_price": 1200,
  "currency": "RUB",
  "from": "07-2025",
  "to": "09-2025",
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "service_name": "Yandex"
}
```

## Формат дат

API принимает даты в формате:

```text
MM-YYYY
```

Например:

```text
07-2025
```

В PostgreSQL даты хранятся как первое число месяца:

```text
2025-07-01
```

## Примеры проверки через PowerShell

### Создание подписки

```powershell
$body = @{
    service_name = "Yandex Plus"
    price = 400
    user_id = "60601fee-2bf1-4721-ae6f-7636e79a0cba"
    start_date = "07-2025"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/subscriptions" -Method Post -ContentType "application/json" -Body $body
```

### Получение списка

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/subscriptions" -Method Get
```

### Подсчёт суммы

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/subscriptions/total-cost?from=07-2025&to=09-2025" -Method Get
```

## Миграции

Миграции находятся в папке:

```text
migrations/
```

Применить миграции:

```powershell
docker compose run --rm migrate
```

При запуске через:

```powershell
docker compose up --build
```

миграции применяются автоматически перед запуском приложения.

## Проверка кода

```powershell
go test ./...
```