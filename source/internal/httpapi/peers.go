package httpapi

import (
	"errors"
	"net/http"

	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/state"
)

func Peers(s *state.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:
			// Return a sorted, independent snapshot.
			WriteJSON(w, http.StatusOK, s.Snapshot())

		case http.MethodPost:
			var peer model.Peer

			// Reuse Week 2 JSON validation.
			if err := DecodeOne(w, r, &peer); err != nil {
				return
			}

			// Validate and add the peer to the registry.
			err := s.Add(peer)

			if err != nil {
				if errors.Is(err, state.ErrDuplicate) {
					ErrorJSON(w, http.StatusConflict, "duplicate_peer")
					return
				}

				ErrorJSON(w, http.StatusBadRequest, "invalid_peer")
				return
			}

			WriteJSON(w, http.StatusCreated, peer)

		default:
			w.Header().Set("Allow", "GET, POST")
			ErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed")
		}
	}
}

func Files(c *state.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			ErrorJSON(w, 405, "method_not_allowed")
			return
		}
		WriteJSON(w, 200, c.Snapshot())
	}
}

var _ = errors.Is
var _ = model.Peer{}
