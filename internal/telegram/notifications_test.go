package telegram

import (
	"strings"
	"testing"

	"github.com/MeguruMacabre/github-release-tracker/internal/model"
)

func TestFormatNotification(t *testing.T) {
	notification := model.Notification{
		ChatID: 10001,
		Release: model.Release{
			Repository: "golang/go",
			Tag:        "go1.27.1",
			URL:        "https://github.com/golang/go/releases/tag/go1.27.1",
		},
	}

	message := formatNotification(notification)

	expected := []string{
		"golang/go",
		"go1.27.1",
		"https://github.com/golang/go/releases/tag/go1.27.1",
	}

	for _, value := range expected {
		if !strings.Contains(message, value) {
			t.Errorf("notification missing %q", value)
		}
	}
}
