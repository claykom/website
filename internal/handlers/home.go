package handlers

import (
	"net/http"

	"github.com/claykom/website/internal/views/pages"
)

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
	response := map[string]string{
		"status": "ok",
	}

	respondWithJSON(w, http.StatusOK, response)
}
