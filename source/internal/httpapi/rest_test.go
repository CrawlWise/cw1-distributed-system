package httpapi

import (
	"example.edu/6cc545/cw1/internal/state"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWeek02_REST(t *testing.T) {
	s := state.NewStatus()
	h := StatusResource(s)
	cases := []struct {
		name, method, body, media string
		code                      int
	}{
		{"get", "GET", "", "", 200}, {"valid", "POST", `{"status":"degraded"}`, "application/json", 200},
		{"charset", "POST", `{"status":"healthy"}`, "application/json; charset=utf-8", 200},
		{"malformed", "POST", `{"status":`, "application/json", 400},
		{"unknown", "POST", `{"status":"healthy","unexpected":1}`, "application/json", 400},
		{"invalid value", "POST", `{"status":"broken"}`, "application/json", 400},
		{"two objects", "POST", `{"status":"healthy"}{}`, "application/json", 400},
		{"null", "POST", `null`, "application/json", 400}, {"array", "POST", `[]`, "application/json", 400},
		{"missing media", "POST", `{}`, "", 415}, {"wrong media", "POST", `{}`, "text/plain", 415},
		{"oversize", "POST", `{"status":"healthy"}` + strings.Repeat(" ", MaxBody), "application/json", 413},
		{"method", "DELETE", "", "", 405},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := s.Get()
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, "/v1/status", strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.media)
			h.ServeHTTP(rr, r)
			if rr.Code != tc.code {
				t.Fatalf("want %d got %d %s", tc.code, rr.Code, rr.Body.String())
			}
			if tc.code >= 400 && s.Get() != before {
				t.Fatal("rejected input changed state")
			}
		})
	}
}
