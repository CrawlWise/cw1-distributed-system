package app

import (
	"context"
	"errors"
	"net/http"

	"example.edu/6cc545/cw1/internal/config"
	"example.edu/6cc545/cw1/internal/httpapi"
	"example.edu/6cc545/cw1/internal/state"
)

func CheckStage(c config.Config) error {
	if c.TCP != "" || c.UDP != "" || c.GRPC != "" || c.Seed != "" {
		return errors.New("Week 2: IPC, RPC and metadata seed are not enabled")
	}
	return nil
}
func Wire(ctx context.Context, c config.Config, mux *http.ServeMux) (func(), error) {
	status := state.NewStatus()
	mux.HandleFunc("/v1/status", httpapi.StatusResource(status))
	return func() {}, nil
}
