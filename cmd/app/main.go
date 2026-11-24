package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MeguruMacabre/github-release-tracker/internal/api"
	"github.com/MeguruMacabre/github-release-tracker/internal/github"
	"github.com/MeguruMacabre/github-release-tracker/internal/storage"
	"github.com/MeguruMacabre/github-release-tracker/internal/telegram"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

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

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	b, err := telegram.New(token, pool)
	if err != nil {
		log.Fatal(err)
	}

	go github.Run(ctx, pool, interval)
	go b.Start(ctx)
	go telegram.RunNotifications(ctx, b, pool, 5*time.Second)

	log.Println("GitHub poller started")
	log.Println("Telegram bot started")
	log.Println("Telegram notifications started")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}

		_, _ = w.Write([]byte("OK"))
	})

	api.Register(mux, pool)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	log.Println("Server started on :8080")

	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server: %v", err)
		}
	}

	stop()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown: %v", err)
	}

	log.Println("Server stopped")
}
