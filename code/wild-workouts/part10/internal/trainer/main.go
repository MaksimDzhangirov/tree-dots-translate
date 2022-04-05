package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MaksimDzhangirov/three-dots/part10/internal/common/genproto/trainer"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/common/logs"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/common/server"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/adapters"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/app"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/app/command"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/app/query"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/domain/hour"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/ports"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

func main() {
	logs.Init()

	ctx := context.Background()
	application := newApplication(ctx)

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		go loadFixtures(application)

		server.RunHTTPServer(func(router chi.Router) http.Handler {
			return ports.HandlerFromMux(
				ports.NewHttpServer(application),
				router,
			)
		})
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := ports.NewGrpcServer(application)
			trainer.RegisterTrainerServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}

func newApplication(ctx context.Context) app.Application {
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	factoryConfig := hour.FactoryConfig{
		MaxWeeksInTheFutureToSet: 6,
		MinUtcHour:               12,
		MaxUtcHour:               20,
	}

	datesRepository := adapters.NewDatesFirestoreRepository(firestoreClient, factoryConfig)

	hourFactory, err := hour.NewFactory(factoryConfig)
	if err != nil {
		panic(err)
	}

	hourRepository := adapters.NewFirestoreHourRepository(firestoreClient, hourFactory)

	return app.Application{
		Commands: app.Commands{
			CancelTraining:       command.NewCancelTrainingHandler(hourRepository),
			ScheduleTraining:     command.NewScheduleTrainingHandler(hourRepository),
			MakeHoursAvailable:   command.NewMakeHoursAvailableHandler(hourRepository),
			MakeHoursUnavailable: command.NewMakeHoursUnavailableHandler(hourRepository),
		},
		Queries: app.Queries{
			HourAvailability:      query.NewHourAvailabilityHandler(hourRepository),
			TrainerAvailableHours: query.NewAvailableHoursHandler(datesRepository),
		},
	}
}
