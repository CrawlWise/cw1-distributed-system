package model

import (
	"encoding/hex"
	"errors"
	"strings"

	"example.edu/6cc545/cw1/internal/config"
)

type Peer struct {
	ID          string `json:"id"`
	HTTPAddress string `json:"http_address"`
	GRPCAddress string `json:"grpc_address"`
	Status      string `json:"status"`
}
type FileMetadata struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	SHA256       string `json:"sha256"`
	ModifiedUnix int64  `json:"modified_unix"`
}

// TODO W03_PEER: implement the contract described in TASKS.md.
func (p Peer) Validate() error {
	// Validate peer identifier.
	if !config.ValidID(p.ID) {
		return errors.New("invalid peer ID")
	}

	// Validate the peer HTTP address.
	if !config.ValidAddress(p.HTTPAddress) {
		return errors.New("invalid peer HTTP address")
	}

	// Validate the peer gRPC address.
	if !config.ValidAddress(p.GRPCAddress) {
		return errors.New("invalid peer gRPC address")
	}

	// Validate recorded peer status.
	switch p.Status {
	case "unknown", "online", "offline":
		return nil
	default:
		return errors.New("invalid peer status")
	}
}

func (f FileMetadata) Validate() error {
	// Validate filename length.
	if f.Name == "" || len(f.Name) > 128 {
		return errors.New("filename must be between 1 and 128 bytes")
	}

	// Reject special directory names.
	if f.Name == "." || f.Name == ".." {
		return errors.New("invalid filename")
	}

	// Prevent directory paths and null bytes.
	if strings.ContainsAny(f.Name, "/\\\x00") {
		return errors.New("filename cannot contain path separators or null bytes")
	}

	// Validate file size.
	if f.Size < 0 {
		return errors.New("file size cannot be negative")
	}

	// Validate modification timestamp.
	if f.ModifiedUnix < 0 {
		return errors.New("modification timestamp cannot be negative")
	}

	// SHA-256 must contain exactly 64 hexadecimal characters.
	if len(f.SHA256) != 64 {
		return errors.New("SHA256 must contain 64 hexadecimal characters")
	}

	// Ensure the hash contains valid hexadecimal characters.
	if _, err := hex.DecodeString(f.SHA256); err != nil {
		return errors.New("SHA256 must contain valid hexadecimal characters")
	}

	return nil
}

var _ = hex.DecodeString
var _ = strings.ContainsAny
var _ = config.ValidID
