package main

import (
	"log"
	"net/http"
	"example.com/pz10-auth/internal/http"
	"example.com/pz10-auth/internal/platform/config"
)

func main() {
	cfg := config.Load()
	mux := router.Build(cfg)
	log.Println("listening on", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, mux))
}