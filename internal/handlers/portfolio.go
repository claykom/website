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
	// Sample data for demonstration
	return &PortfolioHandler{
		projects: []models.Project{
			{
				ID:           "1",
				Title:        "E-commerce Platform",
				Slug:         "ecommerce-platform",
				Description:  "A full-featured e-commerce platform built with Go and React",
				Content:      "This project showcases a complete e-commerce solution with product management, shopping cart functionality, secure payment processing, and order tracking. Built with a Go backend API and a modern React frontend, it demonstrates best practices in full-stack development including RESTful API design, database optimization, and responsive UI design.",
				ImageURL:     "/static/images/ecommerce.jpg",
				ProjectURL:   "https://example.com",
				GithubURL:    "https://github.com/claykom/ecommerce",
				Technologies: []string{"Go", "React", "PostgreSQL", "Docker"},
				Featured:     true,
				CreatedAt:    time.Now().AddDate(0, -6, 0),
				UpdatedAt:    time.Now().AddDate(0, -1, 0),
			},
			{
				ID:           "2",
				Title:        "Task Management API",
				Slug:         "task-management-api",
				Description:  "RESTful API for task management with authentication",
				Content:      "A robust API built with Go, featuring JWT authentication, role-based access control, and comprehensive task management capabilities. The API supports creating, updating, and organizing tasks with tags, priorities, and due dates. It includes automated testing, API documentation with Swagger, and is containerized with Docker for easy deployment.",
				ImageURL:     "/static/images/task-api.jpg",
				ProjectURL:   "https://example.com/tasks",
				GithubURL:    "https://github.com/claykom/task-api",
				Technologies: []string{"Go", "PostgreSQL", "JWT", "REST"},
				Featured:     true,
				CreatedAt:    time.Now().AddDate(0, -3, 0),
				UpdatedAt:    time.Now().AddDate(0, 0, -15),
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
