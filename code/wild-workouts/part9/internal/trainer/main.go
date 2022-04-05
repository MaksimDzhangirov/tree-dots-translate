package main

import (
	"context"
	"fmt"
	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part9/internal/trainer/app"
	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part9/internal/trainer/ports"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part9/internal/trainer/adapters"
	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part9/internal/trainer/domain/hour"
	"github.com/MaksimDzhangirov/three-dots/part9/internal/common/genproto/trainer"
	"github.com/MaksimDzhangirov/three-dots/part9/internal/common/logs"
	"github.com/MaksimDzhangirov/three-dots/part9/internal/common/server"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

func main() {
	logs.Init()

	ctx := context.Background()
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	datesRepository := adapters.NewDatesFirestoreRepository(firestoreClient)

	hourFactory, err := hour.NewFactory(hour.FactoryConfig{
		MaxWeeksInTheFutureToSet: 6,
		MinUtcHour:               12,
		MaxUtcHour:               20,
	})
	if err != nil {
		panic(err)
	}

	hourRepository := adapters.NewFirestoreHourRepository(firestoreClient, hourFactory)

	service := app.NewHourService(datesRepository, hourRepository)

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		go loadFixtures(datesRepository)

		server.RunHTTPServer(func(router chi.Router) http.Handler {
			return ports.HandlerFromMux(
				ports.NewHttpServer(service),
				router,
			)
		})
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := ports.NewGrpcServer(hourRepository)
			trainer.RegisterTrainerServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
