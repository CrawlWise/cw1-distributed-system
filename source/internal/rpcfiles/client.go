package rpcfiles

import (
	"context"
	"strings"
	"time"

	pb "example.edu/6cc545/cw1/api/files/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func Fetch(
	ctx context.Context,
	address, requester string,
	timeout time.Duration,
) (*pb.ListFilesResponse, error) {

	// 1. Respect an already-cancelled caller context.
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}

	// 2. Reject missing or invalid connection addresses.
	if strings.TrimSpace(address) == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"gRPC address is required",
		)
	}

	// 3. Ensure the operation has a finite timeout.
	if timeout <= 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"timeout must be greater than zero",
		)
	}

	// 4. Create an operation context with a deadline.
	opCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 5. Create the gRPC client connection.
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(4096),
			grpc.MaxCallRecvMsgSize(65536),
		),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 6. Create the generated FileService client.
	client := pb.NewFileServiceClient(conn)

	// 7. Invoke the remote ListFiles method.
	response, err := client.ListFiles(
		opCtx,
		&pb.ListFilesRequest{
			RequesterId: requester,
		},
	)
	if err != nil {
		return nil, err
	}

	// 8. Return the remote node's response.
	return response, nil
}
