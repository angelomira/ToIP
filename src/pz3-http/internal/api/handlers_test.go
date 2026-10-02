package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/pz3-http/internal/storage"
)

func TestCreateTask(t *testing.T) {
	store := storage.NewMemoryStore()
	h := NewHandlers(store)

	body := []byte(`{"title":"Купить молоко"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTask(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var task storage.Task
	if err := json.NewDecoder(rr.Body).Decode(&task); err != nil {
		t.Fatal(err)
	}

	if task.ID != 1 || task.Title != "Купить молоко" || task.Done {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	store := storage.NewMemoryStore()
	h := NewHandlers(store)

	body := []byte(`{"title":"ab"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.CreateTask(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

func TestListTasksFilter(t *testing.T) {
	store := storage.NewMemoryStore()
	store.Create("Купить молоко")
	store.Create("Сделать зарядку")

	h := NewHandlers(store)

	req := httptest.NewRequest(http.MethodGet, "/tasks?q=молоко", nil)
	rr := httptest.NewRecorder()

	h.ListTasks(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var tasks []storage.Task
	if err := json.NewDecoder(rr.Body).Decode(&tasks); err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 || tasks[0].Title != "Купить молоко" {
		t.Fatalf("unexpected tasks: %+v", tasks)
	}
}