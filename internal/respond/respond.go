package respond

import (
	"encoding/json"
	"net/http"
)

// JSON writes data as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Error writes a JSON error response: {"error": msg}
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// BindJSON decodes the request body into dst.
// Returns an error if the body is not valid JSON.
func BindJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
