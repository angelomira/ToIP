package notes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler { return &Handler{repo: repo} }

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/stats", h.stats)
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.patch)
	r.Delete("/{id}", h.delete)
	return r
}

func requestContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 5*time.Second)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json_or_title"})
		return
	}
	ctx, cancel := requestContext(r)
	defer cancel()

	note, err := h.repo.Create(ctx, input.Title, input.Content)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "title_already_exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	note, err := h.repo.ByID(ctx, chi.URLParam(r, "id"))
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	skip, _ := strconv.ParseInt(r.URL.Query().Get("skip"), 10, 64)
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if skip < 0 {
		skip = 0
	}

	ctx, cancel := requestContext(r)
	defer cancel()
	items, err := h.repo.List(ctx, query, limit, skip)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	if input.Title == nil && input.Content == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no_fields_to_update"})
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()
	updated, err := h.repo.Update(ctx, chi.URLParam(r, "id"), input.Title, input.Content)
	if errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "title_already_exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	if err := h.repo.Delete(ctx, chi.URLParam(r, "id")); errors.Is(err, ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()
	stats, err := h.repo.GetStats(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database_error"})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
