package command

import (
	"context"
	"time"

	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part12/internal/trainer/domain/hour"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/common/errors"
)

type MakeHoursAvailableHandler struct {
	hourRepo hour.Repository
}

func NewMakeHoursAvailableHandler(hourRepo hour.Repository) MakeHoursAvailableHandler {
	if hourRepo == nil {
		panic("hourRepo is nil")
	}

	return MakeHoursAvailableHandler{hourRepo: hourRepo}
}

func (c MakeHoursAvailableHandler) Handle(ctx context.Context, hours []time.Time) error {
	for _, hourToUpdate := range hours {
		if err := c.hourRepo.UpdateHour(ctx, hourToUpdate, func(h *hour.Hour) (*hour.Hour, error) {
			if err := h.MakeAvailable(); err != nil {
				return nil, err
			}
			return h, nil
		}); err != nil {
			return errors.NewSlugError(err.Error(), "unable-to-update-availability")
		}
	}

	return nil
}
