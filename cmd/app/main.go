package main

import (
	"log"
	"net/http"
	"os"

	"github.com/MeguruMacabre/github-release-tracker/internal/api"
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

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		err := pool.Ping(r.Context())
		if err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Write([]byte("OK"))
	})

	api.Register(mux, pool)

	log.Println("Server started on :8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
