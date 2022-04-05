package hour

import "github.com/pkg/errors"

// Availability - это enum.
//
// Использование структуры вместо `type Availability string` для enum позволяет
// нам полностью управлять возможными значениями.
// Если использовать `type Availability string` вы могли бы создать `Availability("i_can_put_anything_here")`
type Availability struct {
	a string
}

// Каждый тип в Go имеет нулевое значение. В нашем случае - это `Availability{}`.
// Всегда полезно проверить, не равно ли переданное значение нулю!
func (h Availability) IsZero() bool {
	return h == Availability{}
}

var (
	Available         = Availability{"available"}
	NotAvailable      = Availability{"not_available"}
	TrainingScheduled = Availability{"training_scheduled"}
)

var (
	ErrTrainingScheduled   = errors.New("unable to modify hour, because scheduled training")
	ErrNoTrainingScheduled = errors.New("training is not scheduled")
	ErrHourNotAvailable    = errors.New("hour is not available")
)

func (h Hour) IsAvailable() bool {
	return h.availability == Available
}

func (h Hour) HasTrainingScheduled() bool {
	return h.availability == TrainingScheduled
}

func (h *Hour) MakeNotAvailable() error {
	if h.HasTrainingScheduled() {
		return ErrTrainingScheduled
	}

	h.availability = NotAvailable
	return nil
}

func (h *Hour) MakeAvailable() error {
	if h.HasTrainingScheduled() {
		return ErrTrainingScheduled
	}

	h.availability = Available
	return nil
}

func (h *Hour) ScheduleTraining() error {
	if !h.IsAvailable() {
		return ErrHourNotAvailable
	}

	h.availability = TrainingScheduled
	return nil
}

func (h *Hour) CancelTraining() error {
	if !h.HasTrainingScheduled() {
		return ErrNoTrainingScheduled
	}

	h.availability = Available
	return nil
}
