package model

import (
	"strings"
	"testing"
)

func TestWeek03_MetadataValidation(t *testing.T) {
	good := FileMetadata{Name: "notes.txt", Size: 1, SHA256: strings.Repeat("0", 64), ModifiedUnix: 1}
	if e := good.Validate(); e != nil {
		t.Fatal(e)
	}
	cases := []FileMetadata{good, good, good, good}
	cases[0].Name = "../escape"
	cases[1].Size = -1
	cases[2].SHA256 = "not-a-hash"
	cases[3].ModifiedUnix = -1
	for _, f := range cases {
		if f.Validate() == nil {
			t.Fatal("invalid metadata accepted")
		}
	}
}

func TestWeek03_PeerValidation(t *testing.T) {
	good := Peer{
		ID:          "node-b",
		HTTPAddress: "127.0.0.1:8082",
		GRPCAddress: "127.0.0.1:9092",
		Status:      "unknown",
	}

	if err := good.Validate(); err != nil {
		t.Fatalf("valid peer rejected: %v", err)
	}

	tests := []struct {
		name string
		peer Peer
	}{
		{
			name: "invalid ID",
			peer: Peer{
				ID:          "invalid peer",
				HTTPAddress: good.HTTPAddress,
				GRPCAddress: good.GRPCAddress,
				Status:      good.Status,
			},
		},
		{
			name: "invalid HTTP address",
			peer: Peer{
				ID:          good.ID,
				HTTPAddress: "192.168.1.10:8082",
				GRPCAddress: good.GRPCAddress,
				Status:      good.Status,
			},
		},
		{
			name: "invalid gRPC address",
			peer: Peer{
				ID:          good.ID,
				HTTPAddress: good.HTTPAddress,
				GRPCAddress: "127.0.0.1:99999",
				Status:      good.Status,
			},
		},
		{
			name: "invalid status",
			peer: Peer{
				ID:          good.ID,
				HTTPAddress: good.HTTPAddress,
				GRPCAddress: good.GRPCAddress,
				Status:      "connected",
			},
		},
		{
			name: "missing status",
			peer: Peer{
				ID:          good.ID,
				HTTPAddress: good.HTTPAddress,
				GRPCAddress: good.GRPCAddress,
				Status:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.peer.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestWeek03_IPv6PeerValidation(t *testing.T) {
	peer := Peer{
		ID:          "node-ipv6",
		HTTPAddress: "[::1]:8081",
		GRPCAddress: "[::1]:9091",
		Status:      "online",
	}

	if err := peer.Validate(); err != nil {
		t.Fatalf("valid IPv6 loopback peer rejected: %v", err)
	}
}
