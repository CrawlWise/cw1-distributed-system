package httpapi

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "example.edu/6cc545/cw1/api/files/v1"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/state"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testFileService is a configurable gRPC server for HTTP adapter tests.
type testFileService struct {
	pb.UnimplementedFileServiceServer
	list func(context.Context, *pb.ListFilesRequest) (*pb.ListFilesResponse, error)
}

func (s *testFileService) ListFiles(
	ctx context.Context,
	req *pb.ListFilesRequest,
) (*pb.ListFilesResponse, error) {
	return s.list(ctx, req)
}

// startTestGRPC starts a real local gRPC server.
func startTestGRPC(
	t *testing.T,
	list func(context.Context, *pb.ListFilesRequest) (*pb.ListFilesResponse, error),
) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := grpc.NewServer()
	pb.RegisterFileServiceServer(server, &testFileService{
		list: list,
	})

	go func() {
		_ = server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	return listener.Addr().String()
}

// registerTestPeer creates a registry containing node-b.
func registerTestPeer(t *testing.T, address string) *state.Registry {
	t.Helper()

	reg := state.NewRegistry()

	err := reg.Add(model.Peer{
		ID:          "node-b",
		HTTPAddress: "127.0.0.1:8082",
		GRPCAddress: address,
		Status:      "unknown",
	})
	if err != nil {
		t.Fatal(err)
	}

	return reg
}

// requestRemote executes the HTTP handler without starting an HTTP server.
func requestRemote(
	reg *state.Registry,
	method, target string,
	timeout time.Duration,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()

	RemoteFiles(reg, "node-a", timeout).ServeHTTP(rec, req)

	return rec
}

func TestWeek05_RemoteHTTPValidation(t *testing.T) {
	reg := state.NewRegistry()

	tests := []struct {
		name   string
		method string
		target string
		want   int
	}{
		{
			name:   "missing peer ID",
			method: http.MethodGet,
			target: "/v1/remote-files",
			want:   http.StatusBadRequest,
		},
		{
			name:   "malformed peer ID",
			method: http.MethodGet,
			target: "/v1/remote-files?peer_id=invalid%20id",
			want:   http.StatusBadRequest,
		},
		{
			name:   "unknown peer",
			method: http.MethodGet,
			target: "/v1/remote-files?peer_id=node-unknown",
			want:   http.StatusNotFound,
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			target: "/v1/remote-files?peer_id=node-b",
			want:   http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := requestRemote(
				reg,
				tc.method,
				tc.target,
				time.Second,
			)

			if rec.Code != tc.want {
				t.Fatalf(
					"got HTTP %d, want %d: %s",
					rec.Code,
					tc.want,
					rec.Body.String(),
				)
			}
		})
	}
}

func TestWeek05_RemoteHTTPSuccess(t *testing.T) {
	address := startTestGRPC(t,
		func(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
			if req.GetRequesterId() != "node-a" {
				return nil, status.Error(codes.InvalidArgument, "wrong requester")
			}

			return &pb.ListFilesResponse{
				NodeId: "node-b",
				Files: []*pb.FileMetadata{
					{
						Name:         "diagram.txt",
						Size:         25,
						Sha256:       "197756bc66616e3c2a0ffa39792ccbfc157ab7c6efb5b0be7cbc365ee340e613",
						ModifiedUnix: 1700000000,
					},
				},
			}, nil
		},
	)

	reg := registerTestPeer(t, address)

	rec := requestRemote(
		reg,
		http.MethodGet,
		"/v1/remote-files?peer_id=node-b",
		time.Second,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("got HTTP %d: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		NodeID string               `json:"node_id"`
		Files  []model.FileMetadata `json:"files"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	if result.NodeID != "node-b" {
		t.Fatalf("got node ID %q, want node-b", result.NodeID)
	}

	if len(result.Files) != 1 {
		t.Fatalf("got %d files, want 1", len(result.Files))
	}

	file := result.Files[0]

	if file.Name != "diagram.txt" ||
		file.Size != 25 ||
		file.ModifiedUnix != 1700000000 ||
		len(file.SHA256) != 64 {
		t.Fatalf("unexpected file metadata: %+v", file)
	}
}

func TestWeek05_RemoteHTTPEmptyCatalogue(t *testing.T) {
	address := startTestGRPC(t,
		func(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
			return &pb.ListFilesResponse{
				NodeId: "node-b",
				Files:  []*pb.FileMetadata{},
			}, nil
		},
	)

	reg := registerTestPeer(t, address)

	rec := requestRemote(
		reg,
		http.MethodGet,
		"/v1/remote-files?peer_id=node-b",
		time.Second,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("got HTTP %d: %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Files []model.FileMetadata `json:"files"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}

	if result.Files == nil || len(result.Files) != 0 {
		t.Fatalf("expected empty files array, got %s", rec.Body.String())
	}
}

func TestWeek05_RemoteHTTPRPCErrorMapping(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		want int
	}{
		{"invalid argument", codes.InvalidArgument, http.StatusBadRequest},
		{"unavailable", codes.Unavailable, http.StatusServiceUnavailable},
		{"deadline exceeded", codes.DeadlineExceeded, http.StatusGatewayTimeout},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			address := startTestGRPC(t,
				func(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
					return nil, status.Error(tc.code, "simulated remote error")
				},
			)

			reg := registerTestPeer(t, address)

			rec := requestRemote(
				reg,
				http.MethodGet,
				"/v1/remote-files?peer_id=node-b",
				time.Second,
			)

			if rec.Code != tc.want {
				t.Fatalf(
					"got HTTP %d, want %d: %s",
					rec.Code,
					tc.want,
					rec.Body.String(),
				)
			}
		})
	}
}
