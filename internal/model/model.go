package model

import "time"

type Subscription struct {
	ID         int64     `json:"id"`
	ChatID     int64     `json:"chat_id"`
	Repository string    `json:"repository"`
	CreatedAt  time.Time `json:"created_at"`
}

type Release struct {
	ID          int64     `json:"id"`
	Repository  string    `json:"repository"`
	Tag         string    `json:"tag"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
}

type Notification struct {
	ID      int64   `json:"id"`
	ChatID  int64   `json:"chat_id"`
	Release Release `json:"release"`
}
