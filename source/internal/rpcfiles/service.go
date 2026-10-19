package rpcfiles

import (
	"context"

	pb "example.edu/6cc545/cw1/api/files/v1"
	"example.edu/6cc545/cw1/internal/config"
	"example.edu/6cc545/cw1/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Catalog interface{ Snapshot() []model.FileMetadata }
type Service struct {
	pb.UnimplementedFileServiceServer
	NodeID  string
	Catalog Catalog
}

func (s *Service) ListFiles(
	ctx context.Context,
	r *pb.ListFilesRequest,
) (*pb.ListFilesResponse, error) {

	// 1. Check whether the request has been cancelled.
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}

	// 2. Validate the requester ID.
	if r == nil || !config.ValidID(r.GetRequesterId()) {
		return nil, status.Error(
			codes.InvalidArgument,
			"invalid requester_id",
		)
	}

	// 3. Retrieve the shared catalogue snapshot.
	files := s.Catalog.Snapshot()

	// 4. Prepare the response.
	response := &pb.ListFilesResponse{
		NodeId: s.NodeID,
		Files:  make([]*pb.FileMetadata, 0, len(files)),
	}

	// 5. Convert each metadata record.
	for _, file := range files {
		response.Files = append(response.Files, &pb.FileMetadata{
			Name:         file.Name,
			Size:         file.Size,
			Sha256:       file.SHA256,
			ModifiedUnix: file.ModifiedUnix,
		})
	}

	// 6. Return the remote node's metadata.
	return response, nil
}

var _ = config.ValidID
