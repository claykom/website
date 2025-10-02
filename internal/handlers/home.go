package handlers

import (
	"net/http"
	"time"

	"github.com/claykom/website/internal/views/pages"
)

// Application start time for uptime calculation
var startTime = time.Now()

// Home handles the home page
func Home(w http.ResponseWriter, r *http.Request) {
	component := pages.Home()
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

// Health handles health check requests
func Health(w http.ResponseWriter, r *http.Request) {
	// You can add more health checks here (database, external services, etc.)
	status := "ok"
	httpStatus := http.StatusOK

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   getVersion(),
		"uptime":    getUptime(),
	}

	respondWithJSON(w, httpStatus, response)
}

// getVersion returns the application version (you can set this via build flags)
func getVersion() string {
	// This could be set at build time with -ldflags "-X main.version=1.0.0"
	return "1.0.0"
}

// getUptime returns the application uptime
func getUptime() string {
	return time.Since(startTime).String()
}
