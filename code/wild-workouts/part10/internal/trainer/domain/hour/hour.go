package hour

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type Hour struct {
	hour time.Time

	availability Availability
}

type FactoryConfig struct {
	MaxWeeksInTheFutureToSet int
	MinUtcHour               int
	MaxUtcHour               int
}

func (f FactoryConfig) Validate() error {
	var err error

	if f.MaxWeeksInTheFutureToSet < 1 {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MaxWeeksInTheFutureToSet should be greater than 1, but is %d",
				f.MaxWeeksInTheFutureToSet,
			),
		)
	}
	if f.MinUtcHour < 0 || f.MinUtcHour > 24 {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MinUtcHour should be value between 0 and 24, but is %d",
				f.MinUtcHour,
			),
		)
	}
	if f.MaxUtcHour < 0 || f.MaxUtcHour > 24 {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MaxUtcHour should be value between 0 and 24, buti is %d",
				f.MaxUtcHour,
			),
		)
	}

	if f.MinUtcHour > f.MaxUtcHour {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MaxUtcHour (%d) can't be after MinUtcHour (%d)",
				f.MaxUtcHour, f.MinUtcHour,
			),
		)
	}

	return err
}

type Factory struct {
	// лучше будет оставить FactoryConfig приватным полем,
	// благодаря этому мы всегда будем уверены, что наша конфигурация не будет изменена недопустимым образом
	fc FactoryConfig
}

func NewFactory(fc FactoryConfig) (Factory, error) {
	if err := fc.Validate(); err != nil {
		return Factory{}, errors.Wrap(err, "invalid config passed to factory")
	}

	return Factory{fc: fc}, nil
}

func MustNewFactory(fc FactoryConfig) Factory {
	f, err := NewFactory(fc)
	if err != nil {
		panic(err)
	}

	return f
}

func (f Factory) Config() FactoryConfig {
	return f.fc
}

func (f Factory) IsZero() bool {
	return f == Factory{}
}

func (f Factory) NewAvailableHour(hour time.Time) (*Hour, error) {
	if err := f.validateTime(hour); err != nil {
		return nil, err
	}

	return &Hour{
		hour:         hour,
		availability: Available,
	}, nil
}

func (f Factory) NewNotAvailableHour(hour time.Time) (*Hour, error) {
	if err := f.validateTime(hour); err != nil {
		return nil, err
	}

	return &Hour{
		hour:         hour,
		availability: NotAvailable,
	}, nil
}

// UnmarshalHourFromDatabase преобразует Hour из базы данных.
//
// Его следует использовать только для преобразования из базы данных!
// Вы не можете использовать UnmarshalHourFromDatabase в качестве конструктора - он может создать предметную область с недопустимым состоянием!
func (f Factory) UnmarshalHourFromDatabase(hour time.Time, availability Availability) (*Hour, error) {
	if err := f.validateTime(hour); err != nil {
		return nil, err
	}

	if availability.IsZero() {
		return nil, errors.New("empty availability")
	}

	return &Hour{
		hour:         hour,
		availability: availability,
	}, nil
}

var (
	ErrNotFullHour = errors.New("hour should be a full hour")
	ErrPastHour    = errors.New("cannot create hour from past")
)

// Если у вас ошибка с более сложным контекстом,
// хорошей идеей будет выделить её как отдельный тип.
// Нет ничего хуже, чем ошибка "неправильная дата" без знания того, какая дата была передана и какое значение правильное!
type TooDistantDateError struct {
	MaxWeeksInTheFutureToSet int
	ProvidedDate             time.Time
}

func (e TooDistantDateError) Error() string {
	return fmt.Sprintf(
		"schedule can be only set for next %d weeks, provided date: %s",
		e.MaxWeeksInTheFutureToSet,
		e.ProvidedDate,
	)
}

type TooEarlyHourError struct {
	MinUtcHour   int
	ProvidedTime time.Time
}

func (e TooEarlyHourError) Error() string {
	return fmt.Sprintf(
		"too early hour, min UTC hour: %d, provided time: %s",
		e.MinUtcHour,
		e.ProvidedTime,
	)
}

type TooLateHourError struct {
	MaxUtcHour   int
	ProvidedTime time.Time
}

func (e TooLateHourError) Error() string {
	return fmt.Sprintf(
		"too late hour, max UTC hour: %d, provided time: %s",
		e.MaxUtcHour,
		e.ProvidedTime,
	)
}

func (f Factory) validateTime(hour time.Time) error {
	if !hour.Round(time.Hour).Equal(hour) {
		return ErrNotFullHour
	}

	// AddDate лучше, чем Add для добавления дней, потому что в каждый день содержит 24 часа!
	if hour.After(time.Now().AddDate(0, 0, f.fc.MaxWeeksInTheFutureToSet*7)) {
		return TooDistantDateError{
			MaxWeeksInTheFutureToSet: f.fc.MaxWeeksInTheFutureToSet,
			ProvidedDate:             hour,
		}
	}

	currentHour := time.Now().Truncate(time.Hour)
	if hour.Before(currentHour) || hour.Equal(currentHour) {
		return ErrPastHour
	}
	if hour.UTC().Hour() > f.fc.MaxUtcHour {
		return TooLateHourError{
			MaxUtcHour:   f.fc.MaxUtcHour,
			ProvidedTime: hour,
		}
	}
	if hour.UTC().Hour() < f.fc.MinUtcHour {
		return TooEarlyHourError{
			MinUtcHour:   f.fc.MaxUtcHour,
			ProvidedTime: hour,
		}
	}

	return nil
}

func (h *Hour) Time() time.Time {
	return h.hour
}
