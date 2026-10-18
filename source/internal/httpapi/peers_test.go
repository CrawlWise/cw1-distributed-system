package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.edu/6cc545/cw1/internal/state"
)

func TestWeek03_PeerAPI(t *testing.T) {
	h := Peers(state.NewRegistry())
	body := `{"id":"node-b","http_address":"127.0.0.1:8082","grpc_address":"127.0.0.1:9092","status":"unknown"}`
	for _, want := range []int{201, 409} {
		r := httptest.NewRequest("POST", "/v1/peers", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("want %d got %d %s", want, w.Code, w.Body.String())
		}
	}
}

func TestWeek03_PeerAPIGet(t *testing.T) {
	h := Peers(state.NewRegistry())

	body := `{"id":"node-b","http_address":"127.0.0.1:8082","grpc_address":"127.0.0.1:9092","status":"unknown"}`

	post := httptest.NewRequest(
		http.MethodPost,
		"/v1/peers",
		strings.NewReader(body),
	)
	post.Header.Set("Content-Type", "application/json")

	postResponse := httptest.NewRecorder()
	h.ServeHTTP(postResponse, post)

	if postResponse.Code != http.StatusCreated {
		t.Fatalf("POST failed: %d", postResponse.Code)
	}

	get := httptest.NewRequest(http.MethodGet, "/v1/peers", nil)
	getResponse := httptest.NewRecorder()

	h.ServeHTTP(getResponse, get)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getResponse.Code)
	}

	if !strings.Contains(getResponse.Body.String(), "node-b") {
		t.Fatal("GET response does not contain the registered peer")
	}
}

func TestWeek03_PeerAPIInvalid(t *testing.T) {
	h := Peers(state.NewRegistry())

	body := `{"id":"invalid peer","http_address":"127.0.0.1:8082","grpc_address":"127.0.0.1:9092","status":"unknown"}`

	r := httptest.NewRequest(
		http.MethodPost,
		"/v1/peers",
		strings.NewReader(body),
	)
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWeek03_PeerAPIMethodNotAllowed(t *testing.T) {
	h := Peers(state.NewRegistry())

	r := httptest.NewRequest(http.MethodDelete, "/v1/peers", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}

	if w.Header().Get("Allow") != "GET, POST" {
		t.Fatal("incorrect Allow header")
	}
}
