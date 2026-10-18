package app

import (
	"context"
	"encoding/json"
	"errors"
	"example.edu/6cc545/cw1/internal/config"
	"example.edu/6cc545/cw1/internal/httpapi"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/state"
	"io"
	"net/http"
	"os"
)

func loadSeed(c *state.Catalog, path string) error {
	if path == "" {
		return nil
	}
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, 65537))
	if e != nil {
		return e
	}
	if len(b) > 65536 {
		return errors.New("seed exceeds 64 KiB")
	}
	var data []model.FileMetadata
	if e = json.Unmarshal(b, &data); e != nil {
		return e
	}
	return c.Replace(data)
}
func base(c config.Config, mux *http.ServeMux) (*state.Registry, *state.Catalog, error) {
	reg := state.NewRegistry()
	cat := state.NewCatalog()
	if e := loadSeed(cat, c.Seed); e != nil {
		return nil, nil, e
	}
	mux.HandleFunc("/v1/status", httpapi.StatusResource(state.NewStatus()))
	mux.HandleFunc("/v1/peers", httpapi.Peers(reg))
	mux.HandleFunc("/v1/files", httpapi.Files(cat))
	return reg, cat, nil
}

func CheckStage(c config.Config) error {
	if c.TCP != "" || c.UDP != "" || c.GRPC != "" {
		return errors.New("Week 3: IPC and RPC are not enabled")
	}
	return nil
}
func Wire(ctx context.Context, c config.Config, mux *http.ServeMux) (func(), error) {
	_, _, e := base(c, mux)
	return func() {}, e
}
