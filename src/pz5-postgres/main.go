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