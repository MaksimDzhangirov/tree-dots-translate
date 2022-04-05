package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/MaksimDzhangirov/three-dots/internal/common/genproto/trainer"
	"github.com/MaksimDzhangirov/three-dots/internal/common/logs"
	"github.com/MaksimDzhangirov/three-dots/internal/common/server"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strings"
)

func main() {
	logs.Init()

	ctx := context.Background()
	firebaseClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	firebaseDB := db{firebaseClient}

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		go loadFixtures(firebaseDB)

		server.RunHTTPServer(func(router chi.Router) http.Handler {
			return HandlerFromMux(HttpServer{firebaseDB}, router)
		})
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := GrpcServer{db: firebaseDB}
			trainer.RegisterTrainerServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
