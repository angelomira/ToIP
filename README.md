# Практическая работа № 8 — MongoDB + Go

REST-сервис для управления заметками. Проект реализует подключение к MongoDB, коллекцию `notes`, уникальный индекс по `title`, CRUD API, фильтрацию по заголовку, пагинацию, частичное обновление и обработку ошибок. Дополнительно реализован агрегирующий endpoint статистики.

Отчёт находится в папке "`/report/`".

## Требования

- Go 1.21 или новее;
- Docker Desktop с включённым Docker Compose;
- PowerShell (Windows) или терминал;
- curl или Postman (в PowerShell можно использовать `Invoke-RestMethod`).

## Запуск на Windows

Откройте PowerShell в каталоге проекта.

```powershell
Copy-Item .env.example .env
docker compose up -d
docker compose ps
 go mod tidy
 go run ./cmd/api
```

Сервер слушает `http://localhost:8080`. Проверка доступности:

```powershell
Invoke-RestMethod http://localhost:8080/health
```

Ожидаемый результат: `status = ok`.

## API

### Создание заметки

```powershell
$body = @{ title = 'First note'; content = 'Hello Mongo!' } | ConvertTo-Json
$created = Invoke-RestMethod -Method Post -Uri 'http://localhost:8080/api/v1/notes' -ContentType 'application/json' -Body $body
$created
$id = $created.id
```

### Получение списка, поиск и пагинация

```powershell
Invoke-RestMethod 'http://localhost:8080/api/v1/notes?limit=5&skip=0&q=first'
```

### Получение одной заметки

```powershell
Invoke-RestMethod "http://localhost:8080/api/v1/notes/$id"
```

### Частичное обновление

```powershell
$patch = @{ content = 'Updated content' } | ConvertTo-Json
Invoke-RestMethod -Method Patch -Uri "http://localhost:8080/api/v1/notes/$id" -ContentType 'application/json' -Body $patch
```

### Удаление

```powershell
Invoke-WebRequest -Method Delete -Uri "http://localhost:8080/api/v1/notes/$id"
```

Успешное удаление возвращает HTTP 204 No Content.

### Дополнительная статистика (бонус)

```powershell
Invoke-RestMethod 'http://localhost:8080/api/v1/notes/stats'
```

Endpoint возвращает количество документов и среднюю длину поля `content`, вычисленные aggregation pipeline.

## Маршруты

| Метод | Маршрут | Назначение |
|---|---|---|
| GET | `/health` | Проверка HTTP-сервиса |
| POST | `/api/v1/notes` | Создание заметки |
| GET | `/api/v1/notes?q=&limit=&skip=` | Список, поиск и пагинация |
| GET | `/api/v1/notes/stats` | Агрегированная статистика |
| GET | `/api/v1/notes/{id}` | Получение по ObjectID |
| PATCH | `/api/v1/notes/{id}` | Частичное обновление |
| DELETE | `/api/v1/notes/{id}` | Удаление |

## Проверка тестов

Без переменной `MONGO_URI` интеграционный тест пропускается. Для полноценного запуска задайте URI в текущем PowerShell:

```powershell
$env:MONGO_URI = 'mongodb://root:secret@localhost:27017/?authSource=admin'
go test ./...
```

## Структура

```text
pz8-mongo/
├── cmd/api/main.go
├── internal/
│   ├── db/mongo.go
│   └── notes/
│       ├── model.go
│       ├── repo.go
│       ├── handler.go
│       └── repo_test.go
├── .env.example
├── .gitignore
├── docker-compose.yml
├── go.mod
└── README.md
```

## Типовые проблемы

- `connection refused` или timeout: проверьте `docker compose ps` и доступность порта 27017.
- Ошибка аутентификации: проверьте логин `root`, пароль `secret` и параметр `authSource=admin` в URI.
- `duplicate key`: уникальный индекс не позволяет создать две заметки с одинаковым заголовком; API возвращает HTTP 409.
- HTTP 404: проверьте ObjectID и наличие документа.
- Порт 8080 занят: измените `HTTP_ADDR` в `.env`, например на `:8081`, и перезапустите приложение.

## Остановка

```powershell
docker compose down
```

Чтобы также удалить сохранённые данные MongoDB, используйте `docker compose down -v`. Эта команда необратимо удалит volume `mongo_data`.
