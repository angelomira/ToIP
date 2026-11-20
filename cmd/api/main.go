// Package main Notes API server.
//
// @title                      Notes API
// @version                    1.0
// @description                Учебный REST API для заметок (CRUD).
// @contact.name               Backend Course
// @contact.email              example@university.ru
// @BasePath                   /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Введите токен в формате: Bearer <token>
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "example.com/notes-api/docs" // сгенерированные доки swag

	"example.com/notes-api/internal/http/handlers"
	"example.com/notes-api/internal/repo"
)

func main() {
	repository := repo.NewNoteRepoMem()
	h := &handlers.Handler{Repo: repository}

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/notes", h.ListNotes)
		r.Post("/notes", h.CreateNote)
		r.Get("/notes/{id}", h.GetNote)
		r.Patch("/notes/{id}", h.PatchNote)
		r.Delete("/notes/{id}", h.DeleteNote)
	})

	r.Get("/redoc", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>Notes API — ReDoc</title><meta charset="utf-8"/></head>
		<body>
		<redoc spec-url="/docs/swagger.json"></redoc>
		<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
		</body>
		</html>`))
	})

	// Swagger UI: http://localhost:8080/docs/index.html
	r.Get("/docs/*", httpSwagger.WrapHandler)

	log.Println("Server started at :8080")
	log.Println("Swagger UI:  http://localhost:8080/docs/index.html")
	log.Fatal(http.ListenAndServe(":8080", r))
}