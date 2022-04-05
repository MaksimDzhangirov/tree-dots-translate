package command

import (
	"context"
	"time"

	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part10/internal/trainer/domain/hour"
	"github.com/MaksimDzhangirov/three-dots/part10/internal/common/errors"
)

type ScheduleTrainingHandler struct {
	hourRepo hour.Repository
}

func NewScheduleTrainingHandler(hourRepo hour.Repository) ScheduleTrainingHandler {
	if hourRepo == nil {
		panic("nil hourRepo")
	}

	return ScheduleTrainingHandler{hourRepo: hourRepo}
}

func (h ScheduleTrainingHandler) Handle(ctx context.Context, hourToCancel time.Time) error {
	if err := h.hourRepo.UpdateHour(ctx, hourToCancel, func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.ScheduleTraining(); err != nil {
			return nil, err
		}
		return h, nil
	}); err != nil {
		return errors.NewSlugError(err.Error(), "unable-to-update-availability")
	}

	return nil
}
