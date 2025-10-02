package handlers

import (
	"net/http"
	"time"

	"github.com/claykom/website/internal/models"
	"github.com/claykom/website/internal/views/pages"
	"github.com/gorilla/mux"
)

// BlogHandler handles blog-related requests
type BlogHandler struct {
	// In a real application, this would be a database or repository
	posts []models.BlogPost
}

// NewBlogHandler creates a new BlogHandler
func NewBlogHandler() *BlogHandler {
	// Sample data for demonstration
	return &BlogHandler{
		posts: []models.BlogPost{
			{
				ID:          "1",
				Title:       "Getting Started with Go",
				Slug:        "getting-started-with-go",
				Content:     "Go is a powerful programming language created by Google. It's designed for simplicity, efficiency, and excellent concurrency support. Whether you're building web servers, command-line tools, or distributed systems, Go provides the tools you need. In this post, we'll explore the fundamentals of Go and why it's become so popular among developers worldwide.",
				Excerpt:     "Learn the basics of Go programming",
				Author:      "Clay",
				PublishedAt: time.Now().AddDate(0, 0, -7),
				UpdatedAt:   time.Now().AddDate(0, 0, -7),
				Tags:        []string{"go", "programming", "tutorial"},
				Published:   true,
			},
			{
				ID:          "2",
				Title:       "Building Web Servers with Gorilla Mux",
				Slug:        "building-web-servers-gorilla-mux",
				Content:     "Gorilla Mux is a powerful HTTP router and URL matcher for building Go web servers. It provides more features than the standard library's http.ServeMux, including path variables, method-based routing, and middleware support. In this tutorial, we'll build a complete web server using Gorilla Mux and explore best practices for structuring your Go web applications.",
				Excerpt:     "Learn how to build robust web servers",
				Author:      "Clay",
				PublishedAt: time.Now().AddDate(0, 0, -3),
				UpdatedAt:   time.Now().AddDate(0, 0, -3),
				Tags:        []string{"go", "web", "gorilla-mux"},
				Published:   true,
			},
		},
	}
}

// ListPosts returns all published blog posts
func (h *BlogHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	// Filter only published posts
	publishedPosts := make([]models.BlogPost, 0)
	for _, post := range h.posts {
		if post.Published {
			publishedPosts = append(publishedPosts, post)
		}
	}

	component := pages.BlogList(publishedPosts)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

// GetPost returns a single blog post by slug
func (h *BlogHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		http.Error(w, "Slug parameter is required", http.StatusBadRequest)
		return
	}

	// Find post by slug
	for _, post := range h.posts {
		if post.Slug == slug && post.Published {
			component := pages.BlogPost(post)
			if err := component.Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering page", http.StatusInternalServerError)
				return
			}
			return
		}
	}

	http.Error(w, "Blog post not found", http.StatusNotFound)
}

// API handlers (commented out - keeping for reference)
// func (h *BlogHandler) ListPostsAPI(w http.ResponseWriter, r *http.Request) {
// 	publishedPosts := make([]models.BlogPost, 0)
// 	for _, post := range h.posts {
// 		if post.Published {
// 			publishedPosts = append(publishedPosts, post)
// 		}
// 	}
//
// 	respondWithJSON(w, http.StatusOK, map[string]interface{}{
// 		"posts": publishedPosts,
// 		"count": len(publishedPosts),
// 	})
// }
//
// func (h *BlogHandler) GetPostAPI(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	slug := vars["slug"]
//
// 	if slug == "" {
// 		respondWithError(w, http.StatusBadRequest, "Slug parameter is required")
// 		return
// 	}
//
// 	for _, post := range h.posts {
// 		if post.Slug == slug && post.Published {
// 			respondWithJSON(w, http.StatusOK, post)
// 			return
// 		}
// 	}
//
// 	respondWithError(w, http.StatusNotFound, "Blog post not found")
// }
