package app

import (
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/app/command"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/trainer/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CancelTraining   command.CancelTrainingHandler
	ScheduleTraining command.ScheduleTrainingHandler

	MakeHoursAvailable   command.MakeHoursAvailableHandler
	MakeHoursUnavailable command.MakeHoursUnavailableHandler
}

type Queries struct {
	HourAvailability      query.HourAvailabilityHandler
	TrainerAvailableHours query.AvailableHoursHandler
}
