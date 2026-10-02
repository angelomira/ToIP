package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"example.com/pz3-http/internal/storage"
)

type Handlers struct {
	Store *storage.MemoryStore
}

func NewHandlers(store *storage.MemoryStore) *Handlers {
	return &Handlers{Store: store}
}

type createTaskRequest struct {
	Title string `json:"title"`
}

type patchTaskRequest struct {
	Done *bool `json:"done"`
}

// GET /tasks?q=text
func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.Store.List()

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q != "" {
		filtered := make([]*storage.Task, 0)
		lowerQ := strings.ToLower(q)

		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Title), lowerQ) {
				filtered = append(filtered, t)
			}
		}

		tasks = filtered
	}

	JSON(w, http.StatusOK, tasks)
}

// POST /tasks
func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.Contains(ct, "application/json") {
		BadRequest(w, "Content-Type must be application/json")
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid JSON body")
		return
	}

	title := strings.TrimSpace(req.Title)
	if len([]rune(title)) < 3 || len([]rune(title)) > 140 {
		UnprocessableEntity(w, "title must be 3..140 characters")
		return
	}

	task := h.Store.Create(title)
	JSON(w, http.StatusCreated, task)
}

// GET /tasks/{id}
func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(w, r)
	if !ok {
		return
	}

	task, err := h.Store.Get(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			NotFound(w, "task not found")
			return
		}
		Internal(w, "unexpected error")
		return
	}

	JSON(w, http.StatusOK, task)
}

// PATCH /tasks/{id}
func (h *Handlers) PatchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(w, r)
	if !ok {
		return
	}

	var req patchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "invalid JSON body")
		return
	}

	if req.Done == nil {
		BadRequest(w, "field done is required")
		return
	}

	task, err := h.Store.UpdateDone(id, *req.Done)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			NotFound(w, "task not found")
			return
		}
		Internal(w, "unexpected error")
		return
	}

	JSON(w, http.StatusOK, task)
}

// DELETE /tasks/{id}
func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(w, r)
	if !ok {
		return
	}

	if err := h.Store.Delete(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			NotFound(w, "task not found")
			return
		}
		Internal(w, "unexpected error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 || parts[0] != "tasks" {
		NotFound(w, "invalid path")
		return 0, false
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		BadRequest(w, "invalid id")
		return 0, false
	}

	return id, true
}