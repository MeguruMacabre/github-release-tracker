package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/MeguruMacabre/github-release-tracker/internal/model"
)

var repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9-]+/[A-Za-z0-9._-]+$`)

func LatestRelease(ctx context.Context, repository string) (model.Release, error) {
	if !repositoryPattern.MatchString(repository) {
		return model.Release{}, fmt.Errorf("invalid repository: %s", repository)
	}

	url := "https://api.github.com/repos/" + repository + "/releases/latest"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.Release{}, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return model.Release{}, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.Release{}, fmt.Errorf("github API: %s", resp.Status)
	}

	var data struct {
		Tag         string    `json:"tag_name"`
		URL         string    `json:"html_url"`
		PublishedAt time.Time `json:"published_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return model.Release{}, fmt.Errorf("decode release: %w", err)
	}

	if data.Tag == "" || data.URL == "" || data.PublishedAt.IsZero() {
		return model.Release{}, fmt.Errorf("incomplete release data: %s", repository)
	}

	return model.Release{
		Repository:  repository,
		Tag:         data.Tag,
		URL:         data.URL,
		PublishedAt: data.PublishedAt,
	}, nil
}
