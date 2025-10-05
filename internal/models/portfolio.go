package models

import (
	"time"
)

// Project represents a portfolio project
type Project struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	Content      string    `json:"content"`
	ImageURL     string    `json:"image_url"`
	ProjectURL   string    `json:"project_url"`
	GithubURL    string    `json:"github_url"`
	Technologies []string  `json:"technologies"`
	Featured     bool      `json:"featured"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
