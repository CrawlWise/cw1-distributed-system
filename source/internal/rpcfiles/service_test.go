package rpcfiles

import (
	"context"
	pb "example.edu/6cc545/cw1/api/files/v1"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/state"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"strings"
	"testing"
	"time"
)

func TestWeek05_RemoteList(t *testing.T) {
	cat := state.NewCatalog()
	if e := cat.Replace([]model.FileMetadata{{Name: "notes.txt", Size: 1, SHA256: strings.Repeat("a", 64), ModifiedUnix: 1}}); e != nil {
		t.Fatal(e)
	}
	ln, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &Service{NodeID: "node-b", Catalog: cat})
	done := make(chan struct{})
	go func() { defer close(done); _ = s.Serve(ln) }()
	defer func() { s.Stop(); _ = ln.Close(); <-done }()
	got, e := Fetch(context.Background(), ln.Addr().String(), "node-a", time.Second)
	if e != nil {
		t.Fatal(e)
	}
	if got.GetNodeId() != "node-b" || len(got.GetFiles()) != 1 || got.Files[0].GetName() != "notes.txt" {
		t.Fatalf("wrong remote result: %v", got)
	}
	_, e = Fetch(context.Background(), ln.Addr().String(), "", time.Second)
	if status.Code(e) != codes.InvalidArgument {
		t.Fatalf("empty requester: want InvalidArgument got %v", e)
	}
}
func TestWeek05_Unavailable(t *testing.T) {
	ln, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	start := time.Now()
	_, e = Fetch(context.Background(), addr, "node-a", 250*time.Millisecond)
	if e == nil {
		t.Fatal("unavailable peer returned success")
	}
	if status.Code(e) != codes.Unavailable && status.Code(e) != codes.DeadlineExceeded {
		t.Fatalf("unexpected status: %v", e)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("unbounded RPC")
	}
}
func TestWeek05_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, e := (&Service{NodeID: "node-b", Catalog: state.NewCatalog()}).ListFiles(ctx, &pb.ListFilesRequest{RequesterId: "node-a"})
	if status.Code(e) != codes.Canceled {
		t.Fatalf("want Canceled got %v", e)
	}
}
