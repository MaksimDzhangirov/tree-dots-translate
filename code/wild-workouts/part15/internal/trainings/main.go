package main

import (
	"context"
	"net/http"

	"github.com/MaksimDzhangirov/three-dots/part15/internal/common/logs"
	"github.com/MaksimDzhangirov/three-dots/part15/internal/common/server"
	"github.com/MaksimDzhangirov/three-dots/part15/internal/trainings/ports"
	"github.com/MaksimDzhangirov/three-dots/part15/internal/trainings/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	logs.Init()

	ctx := context.Background()

	app, cleanup := service.NewApplication(ctx)
	defer cleanup()

	server.RunHTTPServer(func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpServer(app), router)
	})
}
