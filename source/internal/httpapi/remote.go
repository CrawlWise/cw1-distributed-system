package httpapi

import (
	"net/http"
	"time"

	"example.edu/6cc545/cw1/internal/config"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/rpcfiles"
	"example.edu/6cc545/cw1/internal/state"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RemoteFiles(
	reg *state.Registry,
	nodeID string,
	timeout time.Duration,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Only allow GET requests.
		if r.Method != http.MethodGet {
			ErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}

		// 2. Validate the requested peer ID.
		peerID := r.URL.Query().Get("peer_id")

		if !config.ValidID(peerID) {
			ErrorJSON(w, http.StatusBadRequest, "invalid_peer_id")
			return
		}

		// 3. Look up the peer in the local registry.
		peer, err := reg.Get(peerID)

		if err != nil {
			ErrorJSON(w, http.StatusNotFound, "peer_not_found")
			return
		}

		// 4. Invoke the remote gRPC service.
		response, err := rpcfiles.Fetch(
			r.Context(),
			peer.GRPCAddress,
			nodeID,
			timeout,
		)

		if err != nil {
			// 5. Convert gRPC errors into HTTP responses.
			switch status.Code(err) {
			case codes.InvalidArgument:
				ErrorJSON(w, http.StatusBadRequest, "invalid_argument")

			case codes.Unavailable:
				ErrorJSON(w, http.StatusServiceUnavailable, "remote_unavailable")

			case codes.DeadlineExceeded:
				ErrorJSON(w, http.StatusGatewayTimeout, "remote_timeout")

			case codes.Canceled:
				ErrorJSON(w, http.StatusRequestTimeout, "request_cancelled")

			default:
				ErrorJSON(w, http.StatusBadGateway, "remote_rpc_error")
			}
			return
		}

		// 6. Convert the Protocol Buffer metadata to REST metadata.
		files := make([]model.FileMetadata, 0, len(response.GetFiles()))

		for _, file := range response.GetFiles() {
			files = append(files, model.FileMetadata{
				Name:         file.GetName(),
				Size:         file.GetSize(),
				SHA256:       file.GetSha256(),
				ModifiedUnix: file.GetModifiedUnix(),
			})
		}

		// 7. Return the responding node's identity and files.
		WriteJSON(w, http.StatusOK, struct {
			NodeID string               `json:"node_id"`
			Files  []model.FileMetadata `json:"files"`
		}{
			NodeID: response.GetNodeId(),
			Files:  files,
		})
	}
}
