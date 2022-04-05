package ports

import (
	"context"
	"errors"
	"time"

	"github.com/MaksimDzhangirov/three-dots/part12/internal/common/genproto/trainer"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainer/app"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	trainer.TrainerServiceServer
	app app.Application
}

func NewGrpcServer(application app.Application) GrpcServer {
	return GrpcServer{app: application}
}

func (g GrpcServer) MakeHourAvailable(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
	trainingTime, err := protoTimestampToTime(request.Time)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "unable to parse time")
	}

	if err := g.app.Commands.MakeHoursAvailable.Handle(ctx, []time.Time{trainingTime}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.EmptyResponse{}, nil
}

func (g GrpcServer) ScheduleTraining(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
	trainingTime, err := protoTimestampToTime(request.Time)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "unable to parse time")
	}

	if err := g.app.Commands.ScheduleTraining.Handle(ctx, trainingTime); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.EmptyResponse{}, nil
}

func (g GrpcServer) CancelTraining(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
	trainingTime, err := protoTimestampToTime(request.Time)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "unable to parse time")
	}

	if err := g.app.Commands.CancelTraining.Handle(ctx, trainingTime); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.EmptyResponse{}, nil
}

func (g GrpcServer) IsHourAvailable(ctx context.Context, request *trainer.IsHourAvailableRequest) (*trainer.IsHourAvailableResponse, error) {
	trainingTime, err := protoTimestampToTime(request.Time)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "unable to parse time")
	}

	isAvailable, err := g.app.Queries.HourAvailability.Handle(ctx, trainingTime)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &trainer.IsHourAvailableResponse{IsAvailable: isAvailable}, nil
}

func protoTimestampToTime(timestamp *timestamp.Timestamp) (time.Time, error) {
	if err := timestamp.CheckValid(); err != nil {
		return time.Time{}, errors.New("unable to parse time")
	}

	t := timestamp.AsTime()
	t = t.UTC().Truncate(time.Hour)

	return t, nil
}
