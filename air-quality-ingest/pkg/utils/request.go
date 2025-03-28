package utils

import (
	"encoding/json"
	"net/http"
)

// Decodes the request body into the given payload
func DecodeRequestBody(r *http.Request, payload interface{}) error {
	return json.NewDecoder(r.Body).Decode(payload)
}
