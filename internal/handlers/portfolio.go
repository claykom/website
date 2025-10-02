package handlers

import (
	"net/http"
	"time"

	"github.com/claykom/website/internal/models"
	"github.com/claykom/website/internal/views/pages"
	"github.com/gorilla/mux"
)

// PortfolioHandler handles portfolio-related requests
type PortfolioHandler struct {
	// In a real application, this would be a database or repository
	projects []models.Project
}

// NewPortfolioHandler creates a new PortfolioHandler
func NewPortfolioHandler() *PortfolioHandler {
	// Portfolio project showcasing this website
	return &PortfolioHandler{
		projects: []models.Project{
			{
				ID:           "1",
				Title:        "Personal Website & Portfolio",
				Slug:         "personal-website-portfolio",
				Description:  "A modern, secure Go web application built with security-first design and production-ready deployment",
				Content:      "This website itself serves as a portfolio piece, demonstrating modern Go web development practices. Built with Go 1.25, it features comprehensive security middleware including rate limiting, input validation, and security headers. The application uses Templ for type-safe HTML templating and follows clean architecture principles with a well-organized internal package structure. Container security is implemented through multi-stage Docker builds, non-root user execution, and read-only filesystems. The project includes automated health checks, structured logging, and supports both HTTP and HTTPS deployment with proper TLS configuration. Additional security measures include Content Security Policy headers, XSS protection, HSTS enforcement, and secure static file serving with path traversal protection. The codebase demonstrates Go best practices with comprehensive error handling, graceful shutdown procedures, and environment-based configuration management.",
				ImageURL:     "/static/images/website-portfolio.jpg",
				ProjectURL:   "https://claykom.dev",
				GithubURL:    "https://github.com/claykom/website",
				Technologies: []string{"Go", "Templ", "Docker", "Nginx", "Security"},
				Featured:     true,
				CreatedAt:    time.Now().AddDate(0, 0, -3),
				UpdatedAt:    time.Now(),
			},
		},
	}
}

// ListProjects returns all portfolio projects
func (h *PortfolioHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	component := pages.PortfolioList(h.projects)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

// GetProject returns a single project by slug
func (h *PortfolioHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	if slug == "" {
		http.Error(w, "Slug parameter is required", http.StatusBadRequest)
		return
	}

	// Find project by slug
	for _, project := range h.projects {
		if project.Slug == slug {
			component := pages.ProjectDetail(project)
			if err := component.Render(r.Context(), w); err != nil {
				http.Error(w, "Error rendering page", http.StatusInternalServerError)
				return
			}
			return
		}
	}

	http.Error(w, "Project not found", http.StatusNotFound)
}

// API handlers (commented out - keeping for reference)
// func (h *PortfolioHandler) ListProjectsAPI(w http.ResponseWriter, r *http.Request) {
// 	respondWithJSON(w, http.StatusOK, map[string]interface{}{
// 		"projects": h.projects,
// 		"count":    len(h.projects),
// 	})
// }
//
// func (h *PortfolioHandler) GetProjectAPI(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	slug := vars["slug"]
//
// 	if slug == "" {
// 		respondWithError(w, http.StatusBadRequest, "Slug parameter is required")
// 		return
// 	}
//
// 	for _, project := range h.projects {
// 		if project.Slug == slug {
// 			respondWithJSON(w, http.StatusOK, project)
// 			return
// 		}
// 	}
//
// 	respondWithError(w, http.StatusNotFound, "Project not found")
// }
//
// func (h *PortfolioHandler) ListFeaturedProjectsAPI(w http.ResponseWriter, r *http.Request) {
// 	featuredProjects := make([]models.Project, 0)
// 	for _, project := range h.projects {
// 		if project.Featured {
// 			featuredProjects = append(featuredProjects, project)
// 		}
// 	}
//
// 	respondWithJSON(w, http.StatusOK, map[string]interface{}{
// 		"projects": featuredProjects,
// 		"count":    len(featuredProjects),
// 	})
// }
