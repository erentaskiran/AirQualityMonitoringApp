package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// Helper function to write JSON response
func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error writing JSON response: %v", err)
		// Attempt to write a plain text error if JSON encoding fails
		http.Error(w, `{"error":"Failed to encode JSON response"}`, http.StatusInternalServerError)
	}
}

// Helper function to write JSON error response
func WriteJSONError(w http.ResponseWriter, status int, message string) {
	log.Printf("API Error (%d): %s", status, message)
	WriteJSONResponse(w, status, map[string]string{"error": message})
}
