package httpapi

import (
	"encoding/json"
	"net/http"
)

// WriteJSON is provided: it sets headers before the status and writes JSON.
func WriteJSON(w http.ResponseWriter, code int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "response encoding failed", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, _ = w.Write(append(data, '\n'))
}
func ErrorJSON(w http.ResponseWriter, code int, label string) {
	WriteJSON(w, code, map[string]string{"error": label})
}

func Health(nodeID string) http.HandlerFunc {

	// Return an HTTP handler associated with this node.
	return func(w http.ResponseWriter, r *http.Request) {

		// Only GET requests are supported.
		if r.Method != http.MethodGet {

			// Inform clients which HTTP method is allowed.
			w.Header().Set("Allow", http.MethodGet)

			// Return a JSON error with HTTP status 405.
			ErrorJSON(
				w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
			return
		}

		// Construct the health response for this node.
		response := map[string]string{
			"node_id": nodeID,
			"status":  "ok",
		}

		// Return HTTP 200 with the JSON response.
		WriteJSON(w, http.StatusOK, response)
	}
}
