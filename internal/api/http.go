package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/MeguruMacabre/github-release-tracker/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRequest struct {
	ChatID     int64  `json:"chat_id"`
	Repository string `json:"repository"`
}

func Register(mux *http.ServeMux, pool *pgxpool.Pool) {
	mux.HandleFunc("POST /subscriptions", func(w http.ResponseWriter, r *http.Request) {
		var input SubscriptionRequest

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		input.Repository = strings.TrimSpace(input.Repository)

		if input.ChatID <= 0 || !validRepository(input.Repository) {
			http.Error(w, "Invalid subscription", http.StatusBadRequest)
			return
		}

		subscription, err := storage.AddSubscription(
			r.Context(), pool, input.ChatID, input.Repository,
		)
		if errors.Is(err, storage.ErrSubscriptionExists) {
			http.Error(w, "Subscription already exists", http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(subscription)
	})

	mux.HandleFunc("GET /subscriptions/{chatID}", func(w http.ResponseWriter, r *http.Request) {
		chatID, err := strconv.ParseInt(r.PathValue("chatID"), 10, 64)
		if err != nil || chatID <= 0 {
			http.Error(w, "Invalid chat ID", http.StatusBadRequest)
			return
		}

		subscriptions, err := storage.GetSubscriptions(r.Context(), pool, chatID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(subscriptions)
	})

	mux.HandleFunc("DELETE /subscriptions", func(w http.ResponseWriter, r *http.Request) {
		var input SubscriptionRequest

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		input.Repository = strings.TrimSpace(input.Repository)

		if input.ChatID <= 0 || !validRepository(input.Repository) {
			http.Error(w, "Invalid subscription", http.StatusBadRequest)
			return
		}

		deleted, err := storage.DeleteSubscription(
			r.Context(), pool, input.ChatID, input.Repository,
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if !deleted {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

func validRepository(repository string) bool {
	parts := strings.Split(repository, "/")

	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}
