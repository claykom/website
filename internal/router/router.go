package router

import (
	"net/http"

	"github.com/claykom/website/internal/handlers"
	"github.com/claykom/website/internal/middleware"
	"github.com/gorilla/mux"
)

// New creates and configures a new router with all routes and middleware
func New() *mux.Router {
	r := mux.NewRouter()

	// Initialize handlers
	blogHandler := handlers.NewBlogHandler()
	portfolioHandler := handlers.NewPortfolioHandler()

	// Apply global middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.SecureHeaders)

	// Page routes
	r.HandleFunc("/", handlers.Home).Methods(http.MethodGet)
	r.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	// Blog routes
	r.HandleFunc("/blog", blogHandler.ListPosts).Methods(http.MethodGet)
	r.HandleFunc("/blog/{slug}", blogHandler.GetPost).Methods(http.MethodGet)

	// Portfolio routes
	r.HandleFunc("/portfolio", portfolioHandler.ListProjects).Methods(http.MethodGet)
	r.HandleFunc("/portfolio/{slug}", portfolioHandler.GetProject).Methods(http.MethodGet)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Custom error handlers
	r.NotFoundHandler = http.HandlerFunc(handlers.NotFound)
	r.MethodNotAllowedHandler = http.HandlerFunc(handlers.MethodNotAllowed)

	// API routes (commented out - keeping for reference)
	// api := r.PathPrefix("/api").Subrouter()
	// api.HandleFunc("/blog", blogHandler.ListPostsAPI).Methods(http.MethodGet)
	// api.HandleFunc("/blog/{slug}", blogHandler.GetPostAPI).Methods(http.MethodGet)
	// api.HandleFunc("/portfolio", portfolioHandler.ListProjectsAPI).Methods(http.MethodGet)
	// api.HandleFunc("/portfolio/featured", portfolioHandler.ListFeaturedProjectsAPI).Methods(http.MethodGet)
	// api.HandleFunc("/portfolio/{slug}", portfolioHandler.GetProjectAPI).Methods(http.MethodGet)

	return r
}
