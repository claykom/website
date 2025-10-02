package models

import (
	"time"
)

// BlogPost represents a blog post
type BlogPost struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Content     string    `json:"content"`
	Excerpt     string    `json:"excerpt"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags"`
	Published   bool      `json:"published"`
}
