package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/claykom/website/internal/models"
	"github.com/claykom/website/internal/testutils"
	"github.com/gorilla/mux"
)

func TestBlogHandler_loadMarkdownPosts(t *testing.T) {
	// Create temporary blog directory and files for testing
	tempDir, err := os.MkdirTemp("", "blog_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test blog directory
	blogDir := filepath.Join(tempDir, "blog")
	if err := os.MkdirAll(blogDir, 0755); err != nil {
		t.Fatalf("Failed to create blog dir: %v", err)
	}

	// Create test markdown file
	testContent := `---
title: Test Blog Post
slug: test-blog-post
date: 2024-01-15
description: A test blog post for unit testing
tags: [test, golang]
---

# Test Blog Post

This is a test blog post content.
`

	testFile := filepath.Join(blogDir, "test-post.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Save original working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory so loadMarkdownPosts can find content/blog
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Rename blog dir to content/blog structure
	contentDir := filepath.Join(tempDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create content dir: %v", err)
	}
	if err := os.Rename(blogDir, filepath.Join(contentDir, "blog")); err != nil {
		t.Fatalf("Failed to rename blog dir: %v", err)
	}

	// Test loading posts
	handler := &BlogHandler{}
	err = handler.loadMarkdownPosts()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(handler.posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(handler.posts))
	}

	if len(handler.posts) > 0 {
		post := handler.posts[0]
		if post.Title != "Test Blog Post" {
			t.Errorf("Expected title 'Test Blog Post', got '%s'", post.Title)
		}
		if post.Slug != "test-blog-post" {
			t.Errorf("Expected slug 'test-blog-post', got '%s'", post.Slug)
		}
	}
}

func TestBlogHandler_ListPosts(t *testing.T) {
	handler := &BlogHandler{
		posts: []models.BlogPost{
			{
				ID:          "1",
				Title:       "Test Post 1",
				Slug:        "test-post-1",
				Excerpt:     "First test post",
				Content:     "Content of first post",
				PublishedAt: time.Now(),
				Tags:        []string{"test"},
				Published:   true,
			},
			{
				ID:          "2",
				Title:       "Test Post 2",
				Slug:        "test-post-2",
				Excerpt:     "Second test post",
				Content:     "Content of second post",
				PublishedAt: time.Now(),
				Tags:        []string{"test", "golang"},
				Published:   true,
			},
		},
	}

	req := testutils.NewTestRequest("GET", "/blog", "")
	rr := testutils.NewTestResponseRecorder()

	handler.ListPosts(rr, req)

	// Check response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check that response contains HTML
	body := rr.Body.String()
	if !strings.Contains(body, "<html") || !strings.Contains(body, "</html>") {
		t.Error("Expected response to contain HTML content")
	}
}

func TestBlogHandler_GetPost(t *testing.T) {
	handler := &BlogHandler{
		posts: []models.BlogPost{
			{
				ID:          "1",
				Title:       "Test Post",
				Slug:        "test-post",
				Excerpt:     "A test post",
				Content:     "<h1>Test Content</h1>",
				PublishedAt: time.Now(),
				Tags:        []string{"test"},
				Published:   true,
			},
		},
	}

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		shouldContain  string
	}{
		{
			name:           "existing post",
			slug:           "test-post",
			expectedStatus: http.StatusOK,
			shouldContain:  "<html",
		},
		{
			name:           "non-existing post",
			slug:           "non-existent",
			expectedStatus: http.StatusNotFound,
			shouldContain:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testutils.NewTestRequest("GET", "/blog/"+tt.slug, "")

			// Set up mux vars for the slug parameter
			req = mux.SetURLVars(req, map[string]string{"slug": tt.slug})

			rr := testutils.NewTestResponseRecorder()

			handler.GetPost(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.shouldContain) {
				t.Errorf("Expected response to contain '%s', got: %s", tt.shouldContain, body)
			}
		})
	}
}

func TestPortfolioHandler_ListProjects(t *testing.T) {
	handler := NewPortfolioHandler()

	req := testutils.NewTestRequest("GET", "/portfolio", "")
	rr := testutils.NewTestResponseRecorder()

	handler.ListProjects(rr, req)

	// Check response status
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check that response contains HTML
	body := rr.Body.String()
	if !strings.Contains(body, "<html") || !strings.Contains(body, "</html>") {
		t.Error("Expected response to contain HTML content")
	}

	// Should contain the default portfolio project
	if !strings.Contains(body, "Personal Website") {
		t.Error("Expected response to contain portfolio project content")
	}
}

func TestPortfolioHandler_GetProject(t *testing.T) {
	handler := NewPortfolioHandler()

	tests := []struct {
		name           string
		slug           string
		expectedStatus int
		shouldContain  string
	}{
		{
			name:           "existing project",
			slug:           "personal-website-portfolio",
			expectedStatus: http.StatusOK,
			shouldContain:  "<html",
		},
		{
			name:           "non-existing project",
			slug:           "non-existent",
			expectedStatus: http.StatusNotFound,
			shouldContain:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testutils.NewTestRequest("GET", "/portfolio/"+tt.slug, "")

			// Set up mux vars for the slug parameter
			req = mux.SetURLVars(req, map[string]string{"slug": tt.slug})

			rr := testutils.NewTestResponseRecorder()

			handler.GetProject(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.shouldContain) {
				t.Errorf("Expected response to contain '%s'", tt.shouldContain)
			}
		})
	}
}

func TestNewBlogHandler(t *testing.T) {
	// This test mainly ensures NewBlogHandler doesn't panic
	// and handles missing blog directory gracefully
	handler := NewBlogHandler()

	if handler == nil {
		t.Error("Expected handler to be created")
	}

	// Posts slice should be initialized
	if handler.posts == nil {
		t.Error("Expected posts slice to be initialized")
	}
}

func TestNewPortfolioHandler(t *testing.T) {
	handler := NewPortfolioHandler()

	if handler == nil {
		t.Error("Expected handler to be created")
	}

	// Should have at least one default project
	if len(handler.projects) == 0 {
		t.Error("Expected at least one default project")
	}

	// Check the default project
	project := handler.projects[0]
	if project.Title != "Personal Website & Portfolio" {
		t.Errorf("Expected default project title, got '%s'", project.Title)
	}

	if project.Slug != "personal-website-portfolio" {
		t.Errorf("Expected default project slug, got '%s'", project.Slug)
	}

	if !project.Featured {
		t.Error("Expected default project to be featured")
	}
}

// Test mux URL parameter extraction
func TestMuxURLVars(t *testing.T) {
	// Test that mux.Vars works properly in our test setup
	req := testutils.NewTestRequest("GET", "/blog/test-slug", "")
	req = mux.SetURLVars(req, map[string]string{"slug": "test-slug"})

	vars := mux.Vars(req)
	if slug, ok := vars["slug"]; !ok || slug != "test-slug" {
		t.Errorf("Expected slug 'test-slug', got '%s'", slug)
	}
}

// Benchmark tests
func BenchmarkBlogHandler_ListPosts(b *testing.B) {
	handler := NewBlogHandler()
	req := testutils.NewTestRequest("GET", "/blog", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		handler.ListPosts(rr, req)
	}
}

func BenchmarkPortfolioHandler_ListProjects(b *testing.B) {
	handler := NewPortfolioHandler()
	req := testutils.NewTestRequest("GET", "/portfolio", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := testutils.NewTestResponseRecorder()
		handler.ListProjects(rr, req)
	}
}
