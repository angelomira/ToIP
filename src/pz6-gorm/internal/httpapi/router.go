package httpapi

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func BuildRouter(d *gorm.DB) *chi.Mux {
	r := chi.NewRouter()
	h := NewHandlers(d)

	r.Get("/health", h.Health)
	r.Post("/users", h.CreateUser)
	r.Post("/notes", h.CreateNote)
	r.Get("/notes/{id}", h.GetNoteByID)

	return r
}