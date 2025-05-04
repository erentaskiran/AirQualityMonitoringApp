package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error writing JSON response: %v", err)
		http.Error(w, `{"error":"Failed to encode JSON response"}`, http.StatusInternalServerError)
	}
}

func WriteJSONError(w http.ResponseWriter, status int, message string) {
	log.Printf("API Error (%d): %s", status, message)
	WriteJSONResponse(w, status, map[string]string{"error": message})
}
