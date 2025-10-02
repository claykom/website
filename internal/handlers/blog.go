package handlers

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/claykom/website/internal/models"
	"github.com/claykom/website/internal/views/pages"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
)

// BlogHandler handles blog-related requests
type BlogHandler struct {
	posts []models.BlogPost
}

// NewBlogHandler creates a new BlogHandler and loads markdown posts
func NewBlogHandler() *BlogHandler {
	handler := &BlogHandler{
		posts: []models.BlogPost{},
	}

	// Load posts from markdown files
	if err := handler.loadMarkdownPosts(); err != nil {
		log.Printf("Error loading markdown posts: %v", err)
	}

	return handler
}

// loadMarkdownPosts reads all markdown files from content/blog directory
func (h *BlogHandler) loadMarkdownPosts() error {
	blogDir := "content/blog"

	files, err := os.ReadDir(blogDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(blogDir, file.Name())
		post, err := h.parseMarkdownFile(filePath)
		if err != nil {
			log.Printf("Error parsing %s: %v", filePath, err)
			continue
		}

		h.posts = append(h.posts, post)
	}

	// Sort posts by date (newest first)
	sort.Slice(h.posts, func(i, j int) bool {
		return h.posts[i].PublishedAt.After(h.posts[j].PublishedAt)
	})

	return nil
}

// parseMarkdownFile parses a markdown file with frontmatter
func (h *BlogHandler) parseMarkdownFile(filePath string) (models.BlogPost, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return models.BlogPost{}, err
	}

	// Split frontmatter and content
	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) < 3 {
		return models.BlogPost{}, err
	}

	frontmatter := string(parts[1])
	markdownContent := string(parts[2])

	// Parse frontmatter
	post := models.BlogPost{
		Published: true,
		UpdatedAt: time.Now(),
	}

	scanner := bufio.NewScanner(strings.NewReader(frontmatter))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "title":
			post.Title = value
		case "slug":
			post.Slug = value
			post.ID = value
		case "author":
			post.Author = value
		case "date":
			if t, err := time.Parse("2006-01-02", value); err == nil {
				post.PublishedAt = t
			}
		case "excerpt":
			post.Excerpt = value
		case "tags":
			// Parse tags: [go, programming, tutorial]
			value = strings.Trim(value, "[]")
			tags := strings.Split(value, ",")
			for _, tag := range tags {
				post.Tags = append(post.Tags, strings.TrimSpace(tag))
			}
		}
	}

	// Convert markdown to HTML
	post.Content = h.markdownToHTML(markdownContent)

	return post, nil
}

// markdownToHTML converts markdown to HTML
func (h *BlogHandler) markdownToHTML(md string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.FencedCode
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	doc := p.Parse([]byte(md))
	htmlBytes := markdown.Render(doc, renderer)

	return string(htmlBytes)
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
