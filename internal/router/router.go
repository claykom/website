package router

import (
	"net/http"
	"time"

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

	// Initialize middleware dependencies
	rateLimitStore := middleware.NewRateLimitStore(5 * time.Minute)
	validator := middleware.NewValidator()

	// Apply global middleware in order of importance
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.InputValidation(validator))
	// Rate limit: 100 requests per minute per IP
	r.Use(middleware.RateLimit(rateLimitStore, 100, time.Minute))

	// Page routes
	r.HandleFunc("/", handlers.Home).Methods(http.MethodGet)
	r.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	// Blog routes
	r.HandleFunc("/blog", blogHandler.ListPosts).Methods(http.MethodGet)
	r.HandleFunc("/blog/{slug}", blogHandler.GetPost).Methods(http.MethodGet)

	// Portfolio routes
	r.HandleFunc("/portfolio", portfolioHandler.ListProjects).Methods(http.MethodGet)
	r.HandleFunc("/portfolio/{slug}", portfolioHandler.GetProject).Methods(http.MethodGet)

	// Secure static files handler
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", middleware.SecureStaticHandler(http.Dir("static"))))

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
