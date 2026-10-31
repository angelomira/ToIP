Отчёт находится в папке "`/report/`".

# pz9-auth — регистрация и вход с bcrypt

Учебный HTTP-сервис на Go для практического занятия №9 по дисциплине «Технологии индустриального программирования».

Реализованы:

- `POST /auth/register` — валидация email и пароля, хэширование через bcrypt, сохранение пользователя в PostgreSQL;
- `POST /auth/login` — получение пользователя по email и проверка пароля через `bcrypt.CompareHashAndPassword`;
- `GET /health` — простая проверка доступности сервиса;
- автоматическое создание таблицы `users` через GORM `AutoMigrate`;
- одинаковый ответ `401 invalid_credentials` для неизвестного email и неверного пароля;
- unit-тесты обработчиков с in-memory-хранилищем.

Пароли не возвращаются API и не записываются в открытом виде. Для bcrypt используется cost 12 по умолчанию. Входные пароли длиннее 72 байт отклоняются, поскольку bcrypt имеет ограничение на длину входа.

## Требования

- Go 1.23 или новее;
- PostgreSQL 15+ либо Docker Desktop с Docker Compose;
- доступ к `proxy.golang.org` при первом скачивании Go-модулей.

## Запуск на Windows (PowerShell)

Извлеките архив и откройте PowerShell в каталоге `pz9-auth`.

1. Запустите PostgreSQL через Docker Compose:

```powershell
docker compose up -d db
docker compose ps
```

2. Установите переменные окружения для текущего окна PowerShell:

```powershell
$env:DB_DSN = 'postgres://pz9_user:pz9_password@localhost:5432/pz9_auth?sslmode=disable'
$env:BCRYPT_COST = '12'
$env:APP_ADDR = ':8080'
```

3. Загрузите зависимости, запустите тесты и приложение:

```powershell
go mod tidy
go test ./...
go vet ./...
go run ./cmd/api
```

Сервер будет слушать `http://localhost:8080`. Не закрывайте это окно терминала, пока выполняете HTTP-запросы.

## Запуск на Linux / macOS

```bash
docker compose up -d db
export DB_DSN='postgres://pz9_user:pz9_password@localhost:5432/pz9_auth?sslmode=disable'
export BCRYPT_COST=12
export APP_ADDR=:8080
go mod tidy
go test ./...
go vet ./...
go run ./cmd/api
```

## Проверка API

Команды одинаково подходят для Linux/macOS. В Windows PowerShell используйте `curl.exe`, а не псевдоним `curl`.

Регистрация — ожидается `201 Created`:

```powershell
curl.exe -i -X POST "http://localhost:8080/auth/register" -H "Content-Type: application/json" --data-raw '{"email":"user@example.com","password":"Secret123!"}'
```

Повторная регистрация того же email — ожидается `409 Conflict`:

```powershell
curl.exe -i -X POST "http://localhost:8080/auth/register" -H "Content-Type: application/json" --data-raw '{"email":"user@example.com","password":"AnotherPass123!"}'
```

Вход с правильным паролем — ожидается `200 OK`:

```powershell
curl.exe -i -X POST "http://localhost:8080/auth/login" -H "Content-Type: application/json" --data-raw '{"email":"user@example.com","password":"Secret123!"}'
```

Вход с неправильным паролем — ожидается `401 Unauthorized`:

```powershell
curl.exe -i -X POST "http://localhost:8080/auth/login" -H "Content-Type: application/json" --data-raw '{"email":"user@example.com","password":"WrongPassword123!"}'
```

Неизвестный email также должен вернуть `401 Unauthorized` и тот же JSON с ошибкой `invalid_credentials`.

## Проверка записи в PostgreSQL

```powershell
docker compose exec db psql -U pz9_user -d pz9_auth -c "SELECT id, email, password_hash, created_at, updated_at FROM users ORDER BY id;"
```

В `password_hash` должна находиться bcrypt-строка, начинающаяся с `$2a$` или `$2b$`, а не исходный пароль.

## Тесты

Тесты обработчиков не требуют запущенной PostgreSQL: для них используется временное in-memory-хранилище.

```powershell
go test ./...
go test -cover ./...
go vet ./...
```

## Структура

```text
pz9-auth/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── core/
│   │   └── user.go
│   ├── http/
│   │   └── handlers/
│   │       ├── auth.go
│   │       └── auth_test.go
│   ├── platform/
│   │   └── config/
│   │       └── config.go
│   └── repo/
│       ├── postgres.go
│       └── user_repo.go
├── scripts/
│   ├── test.ps1
│   └── test.sh
├── .env.example
├── .gitignore
├── docker-compose.yml
├── go.mod
└── README.md
```

## Примечания

- `DB_DSN` обязателен; `BCRYPT_COST` по умолчанию равен `12`; `APP_ADDR` по умолчанию равен `:8080`.
- Для учебного запуска в Docker Compose используются локальные пароль и имя пользователя из `docker-compose.yml`. Не используйте эти значения в production.
- JWT намеренно не выдаётся: его добавление относится к следующему практическому занятию №10.
