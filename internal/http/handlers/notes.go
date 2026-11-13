package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"example.com/notes-api/internal/core"
	"example.com/notes-api/internal/repo"
)

type Handler struct {
	Repo *repo.NoteRepoMem
}

// writeError — вспомогательная функция для JSON-ошибок.
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ListNotes godoc
// @Summary      Список заметок
// @Description  Возвращает список заметок с пагинацией и фильтром по заголовку
// @Tags         notes
// @Produce      json
// @Param        page   query  int     false  "Номер страницы"  default(1)
// @Param        limit  query  int     false  "Размер страницы" default(10)
// @Param        q      query  string  false  "Поиск по title"
// @Success      200  {array}  core.Note
// @Header       200  {integer} X-Total-Count "Общее количество"
// @Failure      500  {object}  map[string]string
// @Router       /notes [get]
func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	q := strings.ToLower(r.URL.Query().Get("q"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	all := h.Repo.List()
	filtered := make([]core.Note, 0, len(all))
	for _, n := range all {
		if q == "" || strings.Contains(strings.ToLower(n.Title), q) {
			filtered = append(filtered, n)
		}
	}

	total := len(filtered)
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	_ = json.NewEncoder(w).Encode(filtered[start:end])
}

// CreateNote godoc
// @Summary      Создать заметку
// @Description  Создаёт новую заметку и возвращает её с присвоенным ID
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        input  body      core.NoteCreate  true  "Данные новой заметки"
// @Success      201    {object}  core.Note
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /notes [post]
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var input core.NoteCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(input.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	n := core.Note{
		Title:     input.Title,
		Content:   input.Content,
		CreatedAt: time.Now().UTC(),
	}
	id, err := h.Repo.Create(n)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	n.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(n)
}

// GetNote godoc
// @Summary      Получить заметку по ID
// @Description  Возвращает данные заметки по указанному идентификатору
// @Tags         notes
// @Produce      json
// @Param        id  path      int  true  "Идентификатор заметки"
// @Success      200  {object}  core.Note
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /notes/{id} [get]
func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	n, ok := h.Repo.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n)
}

// PatchNote godoc
// @Summary      Обновить заметку (частично)
// @Description  Обновляет title и/или content заметки
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        id     path      int                true  "ID"
// @Param        input  body      core.NoteUpdate    true  "Поля для обновления"
// @Success      200    {object}  core.Note
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /notes/{id} [patch]
func (h *Handler) PatchNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var input core.NoteUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	n, ok := h.Repo.Update(id, func(n *core.Note) {
		if input.Title != nil {
			n.Title = *input.Title
		}
		if input.Content != nil {
			n.Content = *input.Content
		}
	})
	if !ok {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n)
}

// DeleteNote godoc
// @Summary      Удалить заметку
// @Description  Удаляет заметку по ID
// @Tags         notes
// @Param        id  path  int  true  "ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /notes/{id} [delete]
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !h.Repo.Delete(id) {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}