package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/pz7-redis/internal/cache"
)

func main() {
	c := cache.New("localhost:6379", "Secret123")

	mux := http.NewServeMux()

	// /set?key=...&value=...
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key == "" || value == "" {
			http.Error(w, "key and value required", http.StatusBadRequest)
			return
		}
		err := c.Set(key, value, 10*time.Second) // TTL = 10 сек
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "OK: %s=%s (TTL 10s)", key, value)
	})

	// /get?key=...
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val, err := c.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "VALUE: %s", val)
	})

	// /ttl?key=...
	mux.HandleFunc("/ttl", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		ttl, err := c.TTL(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "TTL: %s", ttl)
	})

	// /del?key=...
	mux.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if err := c.Del(key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "DELETED: %s", key)
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}