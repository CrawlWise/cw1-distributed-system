package app

import (
	"context"
	"errors"
	"example.edu/6cc545/cw1/internal/config"
	"net/http"
)

func CheckStage(c config.Config) error {
	if c.TCP != "" || c.UDP != "" || c.GRPC != "" || c.Seed != "" {
		return errors.New("Week 1: only HTTP is enabled; later-stage flags are reserved")
	}
	return nil
}
func Wire(ctx context.Context, c config.Config, mux *http.ServeMux) (func(), error) {
	return func() {}, nil
}
