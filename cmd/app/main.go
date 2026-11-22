package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MeguruMacabre/github-release-tracker/internal/api"
	"github.com/MeguruMacabre/github-release-tracker/internal/github"
	"github.com/MeguruMacabre/github-release-tracker/internal/storage"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://tracker:tracker@localhost:5433/tracker?sslmode=disable"
	}

	pool, err := storage.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("Connected to PostgreSQL")

	interval := 10 * time.Minute

	if value := os.Getenv("POLL_INTERVAL"); value != "" {
		interval, err = time.ParseDuration(value)
		if err != nil || interval <= 0 {
			log.Fatal("invalid POLL_INTERVAL: ", value)
		}
	}

	go github.Run(context.Background(), pool, interval)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}

		_, _ = w.Write([]byte("OK"))
	})

	api.Register(mux, pool)

	log.Println("Server started on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
