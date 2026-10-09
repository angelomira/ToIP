В репозитории хранится PDF версия отчёта, сделанная в редакторе Obsidian.

Исходный код находится в папке `/src/`.

**Тема:** Подключение к PostgreSQL через database/sql. Выполнение простых запросов (INSERT, SELECT)
**Цели:**
1.        Установить и настроить PostgreSQL локально.
2.        Подключиться к БД из Go с помощью database/sql и драйвера PostgreSQL.
3.        Выполнить параметризованные запросы INSERT и SELECT.
4.        Корректно работать с context, пулом соединений и обработкой ошибок.

## 1. Ход работы

### 1.1. Установка и проверка PostgreSQL

Установил PostgreSQL 16 с сайта EnterpriseDB, задал пароль суперпользователя `postgres`. Убедился, что служба запущена:

![](attachment/ef36e6bd50eddccf49cb8d928b51b9c3.png)

Подключился к серверу (PowerShell):

![](attachment/d2e7254ec7db28fa73291d7a90a5749e.png)
### 1.2. Создание БД и таблицы

Выполним:

```sql
CREATE DATABASE todo;
\c todo

CREATE TABLE IF NOT EXISTS tasks (
    id         SERIAL PRIMARY KEY,
    title      TEXT NOT NULL,
    done       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- добавим одну строку вручную, чтобы проверить работу из psql
INSERT INTO tasks (title) VALUES ('Первая задача из psql');
SELECT * FROM tasks;
```

![](attachment/546a68abf5bf9c61ea1fecdcfc3dfd51.png)
### 1.3. Создание Go-проекта

```powershell
mkdir pz5-postgres
cd pz5-postgres
go mod init pz5-postgres
go get github.com/jackc/pgx/v5/stdlib
go get github.com/joho/godotenv
```

![](attachment/417b8e44e7f18d07e01ea80666d51605.png)

Структура проекта:
```
pz5-postgres/
├── go.mod
├── go.sum
├── .env
├── db.go
├── repository.go
└── main.go
```

Файл `.env`:

```env
DATABASE_URL=postgres://postgres:admin@localhost:5432/todo?sslmode=disable
```
### 1.4. Код проекта
#### 1.4.1. `db.go`

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Настройки пула соединений
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Проверка соединения с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL")
	return db, nil
}
```
#### 1.4.2. `repository.go`

```go
package main

import (
	"context"
	"database/sql"
	"time"
)

// Task — модель для сканирования результатов SELECT
type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{DB: db}
}

// CreateTask — параметризованный INSERT с возвратом id
func (r *Repo) CreateTask(ctx context.Context, title string) (int, error) {
	var id int
	const q = `INSERT INTO tasks (title) VALUES ($1) RETURNING id;`
	err := r.DB.QueryRowContext(ctx, q, title).Scan(&id)
	return id, err
}

// ListTasks — базовый SELECT всех задач
func (r *Repo) ListTasks(ctx context.Context) ([]Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks ORDER BY id;`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ---------- Задание 1: ListDone ----------
func (r *Repo) ListDone(ctx context.Context, done bool) ([]Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE done = $1 ORDER BY id;`
	rows, err := r.DB.QueryContext(ctx, q, done)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ---------- Задание 2: FindByID ----------
func (r *Repo) FindByID(ctx context.Context, id int) (*Task, error) {
	const q = `SELECT id, title, done, created_at FROM tasks WHERE id = $1;`
	var t Task
	err := r.DB.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ---------- Задание 3: CreateMany (через транзакцию) ----------
func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Откат при любой ошибке (в т.ч. панике) до Commit
	defer tx.Rollback()

	const q = `INSERT INTO tasks (title) VALUES ($1);`
	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, title := range titles {
		if _, err := stmt.ExecContext(ctx, title); err != nil {
			return err
		}
	}
	return tx.Commit()
}
```
#### 1.4.3. `main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// .env не обязателен; если файла нет — ошибка игнорируется
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// fallback только для учебного стенда!
		dsn = "postgres://postgres:YOUR_PASSWORD@localhost:5432/todo?sslmode=disable"
	}

	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("openDB error: %v", err)
	}
	defer db.Close()

	repo := NewRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1) INSERT — по одной задаче
	titles := []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"}
	for _, title := range titles {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			log.Fatalf("CreateTask error: %v", err)
		}
		log.Printf("Inserted task id=%d (%s)", id, title)
	}

	// 2) SELECT всех задач
	tasks, err := repo.ListTasks(ctx)
	if err != nil {
		log.Fatalf("ListTasks error: %v", err)
	}
	fmt.Println("=== Tasks ===")
	for _, t := range tasks {
		fmt.Printf("%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	// 3) Задание 1: ListDone
	notDone, err := repo.ListDone(ctx, false)
	if err != nil {
		log.Fatalf("ListDone error: %v", err)
	}
	fmt.Println("\n=== ListDone(false) ===")
	for _, t := range notDone {
		fmt.Printf("%d | %s | done=%v\n", t.ID, t.Title, t.Done)
	}

	// 4) Задание 2: FindByID
	if len(tasks) > 0 {
		t, err := repo.FindByID(ctx, tasks[0].ID)
		if err != nil {
			log.Fatalf("FindByID error: %v", err)
		}
		fmt.Printf("\n=== FindByID(%d) ===\n", t.ID)
		fmt.Printf("ID=%d Title=%q Done=%v CreatedAt=%s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	// 5) Задание 3: CreateMany
	batch := []string{"Задача A", "Задача B", "Задача C"}
	if err := repo.CreateMany(ctx, batch); err != nil {
		log.Fatalf("CreateMany error: %v", err)
	}
	log.Printf("CreateMany inserted %d tasks", len(batch))
}
```
### 1.5. Запуск приложения

Запустим проект:

```
go run .
```

![](attachment/02954b822193f1985548876d14f5e6ae.png)

Проверка через psql:

![](attachment/6521d46c6041b16b3f315d5e36d7b03d.png)

И через новое подключение/клиент:
![](attachment/7b1fc3e4c4e1f2fa4740b0d7fb831936.png)