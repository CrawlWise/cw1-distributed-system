package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.edu/6cc545/cw1/internal/state"
)

func TestWeek02_EmptyBodyPreservesState(t *testing.T) {
	// Arrange: create a status store with a known value.
	s := state.NewStatus()

	if err := s.Set("degraded"); err != nil {
		t.Fatal(err)
	}

	handler := StatusResource(s)

	// Act: send an empty JSON POST request.
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/status",
		strings.NewReader(""),
	)

	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	// Assert: the request must be rejected.
	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected HTTP 400, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	// Assert: rejected input must not modify existing state.
	if got := s.Get(); got != "degraded" {
		t.Fatalf(
			"expected status degraded after rejection, got %q",
			got,
		)
	}
}
