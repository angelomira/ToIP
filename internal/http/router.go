package httpx

import (
	"github.com/go-chi/chi/v5"

	"example.com/notes-api/internal/http/handlers"
)

func NewRouter(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/api/v1/notes", h.CreateNote)
	r.Get("/api/v1/notes", h.ListNotes)
	r.Get("/api/v1/notes/{id}", h.GetNote)
	r.Patch("/api/v1/notes/{id}", h.PatchNote)
	r.Delete("/api/v1/notes/{id}", h.DeleteNote)

	return r
}