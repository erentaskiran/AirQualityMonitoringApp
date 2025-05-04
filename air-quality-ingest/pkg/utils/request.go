package utils

import (
	"encoding/json"
	"net/http"
)

func DecodeRequestBody(r *http.Request, payload interface{}) error {
	return json.NewDecoder(r.Body).Decode(payload)
}
