package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MeguruMacabre/github-release-tracker/internal/api"
	"github.com/MeguruMacabre/github-release-tracker/internal/model"
	"github.com/MeguruMacabre/github-release-tracker/internal/storage"
)

func TestSubscriptionFlow(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	pool, err := storage.Connect(dsn)
	if err != nil {
		t.Fatal(err)
	}

	chatID := time.Now().UnixNano()
	repository := fmt.Sprintf("integration/repo-%d", chatID)

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		_, _ = pool.Exec(cleanupCtx, `
			DELETE FROM notifications
			WHERE release_id IN (
				SELECT id FROM releases WHERE repository = $1
			)
		`, repository)

		_, _ = pool.Exec(
			cleanupCtx,
			"DELETE FROM releases WHERE repository = $1",
			repository,
		)

		_, _ = pool.Exec(
			cleanupCtx,
			"DELETE FROM subscriptions WHERE chat_id = $1",
			chatID,
		)

		pool.Close()
	})

	mux := http.NewServeMux()
	api.Register(mux, pool)

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Создаём подписку через HTTP API
	body := fmt.Sprintf(
		`{"chat_id":%d,"repository":"%s"}`,
		chatID, repository,
	)

	resp, err := client.Post(
		server.URL+"/subscriptions",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST: status = %d", resp.StatusCode)
	}

	// Получаем подписки через HTTP API
	resp, err = client.Get(
		fmt.Sprintf("%s/subscriptions/%d", server.URL, chatID),
	)
	if err != nil {
		t.Fatal(err)
	}

	var subscriptions []model.Subscription
	err = json.NewDecoder(resp.Body).Decode(&subscriptions)
	resp.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	if len(subscriptions) != 1 {
		t.Fatalf("subscriptions count = %d, want 1", len(subscriptions))
	}

	if subscriptions[0].Repository != repository {
		t.Fatalf("wrong repository: %s", subscriptions[0].Repository)
	}

	// Имитируем обнаружение нового релиза
	release := model.Release{
		Repository:  repository,
		Tag:         "v1.0.0",
		URL:         "https://example.com/releases/v1.0.0",
		PublishedAt: time.Now(),
	}

	isNew, err := storage.ProcessRelease(ctx, pool, release)
	if err != nil {
		t.Fatal(err)
	}
	if !isNew {
		t.Fatal("release should be new")
	}

	// Повторная обработка не должна создать дубликат
	isNew, err = storage.ProcessRelease(ctx, pool, release)
	if err != nil {
		t.Fatal(err)
	}
	if isNew {
		t.Fatal("duplicate release detected as new")
	}

	//Проверяем создание уведомления.
	notifications, err := storage.GetPendingNotifications(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}

	var notificationID int64

	for _, n := range notifications {
		if n.ChatID == chatID && n.Release.Repository == repository {
			notificationID = n.ID
			break
		}
	}

	if notificationID == 0 {
		t.Fatal("notification was not created")
	}

	//Имитируем подтверждение отправки.
	sent, err := storage.MarkNotificationSent(ctx, pool, notificationID)
	if err != nil {
		t.Fatal(err)
	}
	if !sent {
		t.Fatal("notification was not marked as sent")
	}

	//Удаляем подписку через HTTP API.
	req, err := http.NewRequest(
		http.MethodDelete,
		server.URL+"/subscriptions",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE: status = %d", resp.StatusCode)
	}

	t.Log("Subscription flow completed successfully")
}
