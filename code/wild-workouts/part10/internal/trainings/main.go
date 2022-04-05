package main

import (
	"context"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	grpcClient "github.com/MaksimDzhangirov/three-dots/part10/internal/common/client"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/common/logs"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/common/server"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainings/adapters"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainings/app"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainings/ports"
	"github.com/go-chi/chi/v5"
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

	trainingsRepository := adapters.NewTrainingsFirestoreRepository(client)
	trainerGrpc := adapters.NewTrainerGrpc(trainerClient)
	usersGrpc := adapters.NewUsersGrpc(usersClient)

	trainingsService := app.NewTrainingsService(trainingsRepository, trainerGrpc, usersGrpc)

	server.RunHTTPServer(func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpServer(trainingsService), router)
	})
}
