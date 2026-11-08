package main

import (
	"log"
	"net/http"

	httpx "example.com/notes-api/internal/http"
	"example.com/notes-api/internal/http/handlers"
	"example.com/notes-api/internal/repo"
)

func main() {
	noteRepo := repo.NewNoteRepoMem()
	h := &handlers.Handler{Repo: noteRepo}
	r := httpx.NewRouter(h)

	log.Println("Server started at :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}