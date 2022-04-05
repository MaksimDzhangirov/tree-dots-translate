package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/MaksimDzhangirov/three-dots/internal/common/logs"
	"github.com/MaksimDzhangirov/three-dots/internal/common/server"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"

	grpcClient "github.com/MaksimDzhangirov/three-dots/internal/common/client"
)

func main() {
	logs.Init()

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	trainerClient, closeTrainerClient, err := grpcClient.NewTrainerClient()
	if err != nil {
		panic(err)
	}
	defer closeTrainerClient()

	usersClient, closeUsersClient, err := grpcClient.NewUsersClient()
	if err != nil {
		panic(err)
	}
	defer closeUsersClient()

	firebaseDB := db{client}

	server.RunHTTPServer(func(router chi.Router) http.Handler {
		return HandlerFromMux(HttpServer{firebaseDB, trainerClient, usersClient}, router)
	})
}