package handlers

import (
	"encoding/json"
	"net/http"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct{}

// RespondJSON sends a JSON response
func (h *BaseHandler) RespondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// RespondError sends an error response
func (h *BaseHandler) RespondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
