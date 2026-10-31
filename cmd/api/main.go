package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/pz8-mongo/internal/db"
	"example.com/pz8-mongo/internal/notes"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Loading .env is optional: environment variables take precedence.
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found; using process environment and defaults")
	}

	uri := getenv("MONGO_URI", "mongodb://root:secret@localhost:27017/?authSource=admin")
	dbName := getenv("MONGO_DB", "pz8")
	addr := getenv("HTTP_ADDR", ":8080")

	deps, err := db.ConnectMongo(context.Background(), uri, dbName)
	if err != nil {
		log.Fatal("MongoDB connection failed: ", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := deps.Client.Disconnect(ctx); err != nil {
			log.Println("MongoDB disconnect error:", err)
		}
	}()

	repo, err := notes.NewRepo(deps.Database)
	if err != nil {
		log.Fatal("notes repository initialization failed: ", err)
	}
	handler := notes.NewHandler(repo)

	router := chi.NewRouter()
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	router.Mount("/api/v1/notes", handler.Routes())

	server := &http.Server{Addr: addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Printf("HTTP server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error: ", err)
		}
	}()
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Println("HTTP shutdown error:", err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
