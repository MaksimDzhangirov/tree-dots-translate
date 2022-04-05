package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainer/domain/hour"
)

type MemoryHourRepository struct {
	hours map[time.Time]hour.Hour
	lock  *sync.RWMutex

	hourFactory hour.Factory
}

func NewMemoryHourRepository(hourFactory hour.Factory) *MemoryHourRepository {
	if hourFactory.IsZero() {
		panic("missing hourFactory")
	}

	return &MemoryHourRepository{
		hours:       map[time.Time]hour.Hour{},
		lock:        &sync.RWMutex{},
		hourFactory: hourFactory,
	}
}

func (m MemoryHourRepository) GetHour(_ context.Context, hourTime time.Time) (*hour.Hour, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return m.getOrCreateHour(hourTime)
}

func (m MemoryHourRepository) getOrCreateHour(hourTime time.Time) (*hour.Hour, error) {
	currentHour, ok := m.hours[hourTime]
	if !ok {
		return m.hourFactory.NewNotAvailableHour(hourTime)
	}

	// мы храним часы не как указатели, а как значения
	// благодаря этому, мы уверены, что никто не сможет изменить Hour, не используя UpdateHour
	return &currentHour, nil
}

func (m MemoryHourRepository) UpdateHour(
	_ context.Context,
	hourTime time.Time,
	updateFn func(h *hour.Hour) (*hour.Hour, error),
) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	currentHour, err := m.getOrCreateHour(hourTime)
	if err != nil {
		return err
	}

	updatedHour, err := updateFn(currentHour)
	if err != nil {
		return err
	}

	m.hours[hourTime] = *updatedHour

	return nil
}
