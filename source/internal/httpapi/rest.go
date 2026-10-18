package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"example.edu/6cc545/cw1/internal/state"
)

const MaxBody = 4096

// DecodeOne enforces a single bounded JSON object. No state is mutated here.
func DecodeOne(w http.ResponseWriter, r *http.Request, dst any) error {
	// Step 1: Validate the request media type.
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &InputError{
			Status: http.StatusUnsupportedMediaType,
			Label:  "unsupported_media_type",
		}
	}

	// Step 2: Limit how many bytes can be read.
	r.Body = http.MaxBytesReader(w, r.Body, MaxBody)

	// Step 3: Read the bounded request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return &InputError{
				Status: http.StatusRequestEntityTooLarge,
				Label:  "body_too_large",
			}
		}

		return &InputError{
			Status: http.StatusBadRequest,
			Label:  "invalid_input",
		}
	}

	// Step 4: Ensure the top-level JSON value is an object.
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return &InputError{
			Status: http.StatusBadRequest,
			Label:  "invalid_json_object",
		}
	}

	// Step 5: Decode JSON and reject unknown fields.
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return &InputError{
			Status: http.StatusBadRequest,
			Label:  "invalid_json",
		}
	}

	// Step 6: Reject a second JSON value or trailing junk.
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return &InputError{
			Status: http.StatusBadRequest,
			Label:  "multiple_json_values",
		}
	}

	return nil
}

type InputError struct {
	Status int
	Label  string
}

func (e *InputError) Error() string { return e.Label }
func RejectInput(w http.ResponseWriter, e error) {
	var in *InputError
	if errors.As(e, &in) {
		ErrorJSON(w, in.Status, in.Label)
		return
	}
	ErrorJSON(w, 400, "invalid_input")
}

func StatusResource(s *state.Status) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:
			// Read the current status without changing it.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(map[string]string{
				"status": s.Get(),
			})

		case http.MethodPost:
			// Decode into a temporary request object.
			var request struct {
				Status string `json:"status"`
			}

			// Validate the incoming JSON before changing state.
			if err := DecodeOne(w, r, &request); err != nil {
				RejectInput(w, err)
				return
			}

			// Apply the change only after successful decoding.
			if err := s.Set(request.Status); err != nil {
				ErrorJSON(w, http.StatusBadRequest, "invalid_status")
				return
			}

			// Return the updated status.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(map[string]string{
				"status": s.Get(),
			})

		default:
			// Reject unsupported HTTP methods.
			w.Header().Set("Allow", "GET, POST")
			ErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	}
}
