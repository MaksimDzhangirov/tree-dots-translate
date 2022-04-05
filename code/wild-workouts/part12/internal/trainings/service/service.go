package service

import (
	"context"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainings/app/query"
	"os"

	"cloud.google.com/go/firestore"
	grpcClient "github.com/MaksimDzhangirov/three-dots/part12/internal/common/client"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainings/adapters"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainings/app"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainings/app/command"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	trainerClient, closeTrainerClient, err := grpcClient.NewTrainerClient()
	if err != nil {
		panic(err)
	}

	usersClient, closeUsersClient, err := grpcClient.NewUsersClient()
	if err != nil {
		panic(err)
	}
	trainerGrpc := adapters.NewTrainerGrpc(trainerClient)
	usersGrpc := adapters.NewUsersGrpc(usersClient)

	return newApplication(ctx, trainerGrpc, usersGrpc),
		func() {
			_ = closeTrainerClient()
			_ = closeUsersClient()
		}
}

func NewComponentTestApplication(ctx context.Context) app.Application {
	return newApplication(ctx, TrainerServiceMock{}, UserServiceMock{})
}

func newApplication(ctx context.Context, trainerGrpc command.TrainerService, usersGrpc command.UserService) app.Application {
	client, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	trainingsRepository := adapters.NewTrainingsFirestoreRepository(client)

	return app.Application{
		Commands: app.Commands{
			ApproveTrainingReschedule: command.NewApproveTrainingRescheduleHandler(trainingsRepository, usersGrpc, trainerGrpc),
			CancelTraining:            command.NewCancelTrainingHandler(trainingsRepository, usersGrpc, trainerGrpc),
			RejectTrainingReschedule:  command.NewRejectTrainingRescheduleHandler(trainingsRepository),
			RescheduleTraining:        command.NewRescheduleTrainingHandler(trainingsRepository, usersGrpc, trainerGrpc),
			RequestTrainingReschedule: command.NewRequestTrainingRescheduleHandler(trainingsRepository),
			ScheduleTraining:          command.NewScheduleTrainingHandler(trainingsRepository, usersGrpc, trainerGrpc),
		},
		Queries: app.Queries{
			AllTrainings:     query.NewAllTrainingsHandler(trainingsRepository),
			TrainingsForUser: query.NewTrainingsForUserHandler(trainingsRepository),
		},
	}
}
