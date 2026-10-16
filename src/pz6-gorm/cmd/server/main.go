package main

import (
	"log"
	"net/http"

	"example.com/pz6-gorm/internal/db"
	httpapi "example.com/pz6-gorm/internal/httpapi"
	"example.com/pz6-gorm/internal/models"
)

func main() {
	d := db.Connect()

	// AutoMigrate создаст таблицы users, notes, tags, note_tags
	if err := d.AutoMigrate(&models.User{}, &models.Note{}, &models.Tag{}); err != nil {
		log.Fatal("migrate:", err)
	}

	r := httpapi.BuildRouter(d)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}