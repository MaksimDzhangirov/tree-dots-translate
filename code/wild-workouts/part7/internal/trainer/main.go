package main

import (
	"context"
	"fmt"
	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part7/internal/trainer/domain/hour"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/MaksimDzhangirov/three-dots/part7/internal/common/genproto/trainer"
	"github.com/MaksimDzhangirov/three-dots/part7/internal/common/logs"
	"github.com/MaksimDzhangirov/three-dots/part7/internal/common/server"
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

	firebaseDB := db{firestoreClient}

	hourFactory, err := hour.NewFactory(hour.FactoryConfig{
		MaxWeeksInTheFutureToSet: 6,
		MinUtcHour:               12,
		MaxUtcHour:               20,
	})
	if err != nil {
		panic(err)
	}

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		go loadFixtures(firebaseDB)

		server.RunHTTPServer(func(router chi.Router) http.Handler {
			return HandlerFromMux(
				HttpServer{
					firebaseDB,
					NewFirestoreHourRepository(firestoreClient, hourFactory),
				},
				router,
			)
		})
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := GrpcServer{hourRepository: NewFirestoreHourRepository(firestoreClient, hourFactory)}
			trainer.RegisterTrainerServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
