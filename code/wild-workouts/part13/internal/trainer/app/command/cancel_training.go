package command

import (
	"context"
	"time"

	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part13/internal/trainer/domain/hour"
	"github.com/MaksimDzhangirov/three-dots/part13/internal/common/errors"
)

type CancelTrainingHandler struct {
	hourRepo hour.Repository
}

func NewCancelTrainingHandler(hourRepo hour.Repository) CancelTrainingHandler {
	if hourRepo == nil {
		panic("nil hourRepo")
	}

	return CancelTrainingHandler{hourRepo: hourRepo}
}

func (h CancelTrainingHandler) Handle(ctx context.Context, hourToCancel time.Time) error {
	if err := h.hourRepo.UpdateHour(ctx, hourToCancel, func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.CancelTraining(); err != nil {
			return nil, err
		}
		return h, nil
	}); err != nil {
		return errors.NewSlugError(err.Error(), "unable-to-update-availability")
	}

	return nil
}
