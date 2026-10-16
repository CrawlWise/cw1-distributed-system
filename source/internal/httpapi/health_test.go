package httpapi

import (
	"encoding/json"
	"net/http" // Included the http to allow me carry out a new test
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSmoke_JSONHelper(t *testing.T) {
	r := httptest.NewRecorder()
	WriteJSON(r, 200, map[string]string{"hello": "world"})
	if r.Code != 200 || !json.Valid(r.Body.Bytes()) {
		t.Fatal("provided JSON helper failed")
	}
}
func TestWeek01_Health(t *testing.T) {
	for _, id := range []string{"node-a", "node-b"} {
		rr := httptest.NewRecorder()
		Health(id).ServeHTTP(rr, httptest.NewRequest("GET", "/healthz", nil))
		var got struct {
			Status string `json:"status"`
			NodeID string `json:"node_id"`
		}
		if e := json.Unmarshal(rr.Body.Bytes(), &got); e != nil {
			t.Fatal(e)
		}
		if rr.Code != 200 || got.Status != "ok" || got.NodeID != id {
			t.Fatalf("incorrect health: %d %s", rr.Code, rr.Body.String())
		}
		if !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json") {
			t.Fatal("missing JSON content type")
		}
	}
}
func TestWeek01_HealthWrongMethod(t *testing.T) {
	rr := httptest.NewRecorder()
	Health("node-a").ServeHTTP(rr, httptest.NewRequest("POST", "/healthz", nil))
	if rr.Code != 405 {
		t.Fatalf("want 405 got %d", rr.Code)
	}
}

func TestWeek01_HealthAllowHeader(t *testing.T) {
	// Create a simulated HTTP POST request.
	req := httptest.NewRequest(
		http.MethodPost,
		"/healthz",
		nil,
	)

	// Record the HTTP response.
	recorder := httptest.NewRecorder()

	// Execute the health handler.
	Health("node-a").ServeHTTP(recorder, req)

	// Verify that unsupported methods return HTTP 405.
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected HTTP 405, got %d",
			recorder.Code,
		)
	}

	// Verify that the server identifies the allowed method.
	if got := recorder.Header().Get("Allow"); got != "GET" {
		t.Errorf(
			"expected Allow header GET, got %q",
			got,
		)
	}

	// Verify the JSON error response.
	var response map[string]string

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if response["error"] != "method not allowed" {
		t.Errorf(
			"unexpected error response: %v",
			response,
		)
	}
}
