package adapters_test

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainer/adapters"

	"cloud.google.com/go/firestore"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainer/domain/hour"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	repositories := createRepositories(t)

	for i := range repositories {
		// Когда вы перебираете срез и позже используете повторяющееся значение в goroutine (здесь из-за t.Parallel ()),
		// вам нужно всегда создавать переменную в теле цикла!
		// Подробнее здесь: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		r := repositories[i]

		t.Run(r.Name, func(t *testing.T) {
			// Всегда полезно создавать все не unit тесты, чтобы они могли работать параллельно.
			// Благодаря этому ваши тесты всегда будут быстро выполняться, и вы не будете бояться добавлять тесты из-за замедления.
			t.Parallel()

			t.Run("testUpdateHour", func(t *testing.T) {
				t.Parallel()
				testUpdateHour(t, r.Repository)
			})
			t.Run("testUpdateHour_parallel", func(t *testing.T) {
				t.Parallel()
				testUpdateHour_parallel(t, r.Repository)
			})
			t.Run("testHourRepository_update_existing", func(t *testing.T) {
				t.Parallel()
				testHourRepository_update_existing(t, r.Repository)
			})
			t.Run("testUpdateHour_rollback", func(t *testing.T) {
				t.Parallel()
				testUpdateHour_rollback(t, r.Repository)
			})
		})
	}
}

type Repository struct {
	Name       string
	Repository hour.Repository
}

func createRepositories(t *testing.T) []Repository {
	return []Repository{
		{
			Name:       "Firebase",
			Repository: newFirebaseRepository(t, context.Background()),
		},
		{
			Name:       "MySQL",
			Repository: newMySQLRepository(t),
		},
		{
			Name:       "memory",
			Repository: adapters.NewMemoryHourRepository(testHourFactory),
		},
	}
}

func testUpdateHour(t *testing.T, repository hour.Repository) {
	t.Helper()
	ctx := context.Background()

	testCases := []struct {
		Name       string
		CreateHour func(*testing.T) *hour.Hour
	}{
		{
			Name: "available_hour",
			CreateHour: func(t *testing.T) *hour.Hour {
				return newValidAvailableHour(t)
			},
		},
		{
			Name: "not_available_hour",
			CreateHour: func(t *testing.T) *hour.Hour {
				h := newValidAvailableHour(t)
				require.NoError(t, h.MakeNotAvailable())

				return h
			},
		},
		{
			Name: "hour_with_training",
			CreateHour: func(t *testing.T) *hour.Hour {
				h := newValidAvailableHour(t)
				require.NoError(t, h.ScheduleTraining())

				return h
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newHour := tc.CreateHour(t)

			err := repository.UpdateHour(ctx, newHour.Time(), func(_ *hour.Hour) (*hour.Hour, error) {
				// UpdateHour выдаёт нам существующий/новый *hour.Hour,
				// но мы игнорируем этот час и сохраняем результат `CreateHour`
				// мы можем проверить значение этого часа позднее в assertHourInRepository
				return newHour, nil
			})
			require.NoError(t, err)

			assertHourInRepository(ctx, t, repository, newHour)
		})
	}
}

func testUpdateHour_parallel(t *testing.T, repository hour.Repository) {
	if _, ok := repository.(*adapters.FirestoreHourRepository); ok {
		// todo - enable after fix of https://github.com/googleapis/google-cloud-go/issues/2604
		t.Skip("because of emulator bug, it's not working in Firebase")
	}

	t.Helper()
	ctx := context.Background()

	hourTime := newValidHourTime()

	// мы добавляем доступный час
	err := repository.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.MakeAvailable(); err != nil {
			return nil, err
		}
		return h, nil
	})
	require.NoError(t, err)

	workersCount := 20
	workersDone := sync.WaitGroup{}
	workersDone.Add(workersCount)

	// закрытие startWorkers разблокирует всех workerы сразу,
	// благодаря этому будет больше возможности получить состояние гонки
	startWorkers := make(chan struct{})
	// если обучение было успешно запланировано, номер workerа отправляется в этот канал
	trainingsScheduled := make(chan int, workersCount)

	// мы пытаемся создать условие гонки, на практике только один worker должен быть в состоянии завершить транзакцию
	for worker := 0; worker < workersCount; worker++ {
		workerNum := worker

		go func() {
			defer workersDone.Done()
			<-startWorkers

			schedulingTraining := false

			err := repository.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
				// тренировка уже запланирована, ничего делать не нужно
				if h.HasTrainingScheduled() {
					return h, nil
				}
				// тренировка еще не запланирована, так что давайте попробуем это сделать
				if err := h.ScheduleTraining(); err != nil {
					return nil, err
				}

				schedulingTraining = true

				return h, nil
			})

			if schedulingTraining && err == nil {
				// обучение планируется только в том случае, если UpdateHour не вернул ошибку
				trainingsScheduled <- workerNum
			}
		}()
	}

	close(startWorkers)

	// ждём, когда все workerы закончат работу
	workersDone.Wait()
	close(trainingsScheduled)

	var workersScheduledTraining []int

	for workerNum := range trainingsScheduled {
		workersScheduledTraining = append(workersScheduledTraining, workerNum)
	}

	assert.Len(t, workersScheduledTraining, 1, "only one worker should schedule training")
}

func testUpdateHour_rollback(t *testing.T, repository hour.Repository) {
	t.Helper()
	ctx := context.Background()

	hourTime := newValidHourTime()

	err := repository.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
		require.NoError(t, h.MakeAvailable())
		return h, nil
	})

	err = repository.UpdateHour(ctx, hourTime, func(h *hour.Hour) (*hour.Hour, error) {
		assert.True(t, h.IsAvailable())
		require.NoError(t, h.MakeNotAvailable())

		return h, errors.New("something went wrong")
	})
	require.Error(t, err)

	persistedHour, err := repository.GetHour(ctx, hourTime)
	require.NoError(t, err)

	assert.True(t, persistedHour.IsAvailable(), "availability change was persisted, not rolled back")
}

// testHourRepository_update_existing - тестовая последовательность создания нового часа и его обновления.
func testHourRepository_update_existing(t *testing.T, repository hour.Repository) {
	t.Helper()
	ctx := context.Background()

	testHour := newValidAvailableHour(t)

	err := repository.UpdateHour(ctx, testHour.Time(), func(_ *hour.Hour) (*hour.Hour, error) {
		return testHour, nil
	})
	require.NoError(t, err)
	assertHourInRepository(ctx, t, repository, testHour)

	var expectedHour *hour.Hour
	err = repository.UpdateHour(ctx, testHour.Time(), func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.ScheduleTraining(); err != nil {
			return nil, err
		}
		expectedHour = h
		return h, nil
	})
	require.NoError(t, err)

	assertHourInRepository(ctx, t, repository, expectedHour)
}

func TestNewDateDTO(t *testing.T) {
	testCases := []struct {
		Time             time.Time
		ExpectedDateTime time.Time
	}{
		{
			Time:             time.Date(3333, 1, 1, 0, 0, 0, 0, time.UTC),
			ExpectedDateTime: time.Date(3333, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			// мы сохраняем дату в UTC
			// в UTC будет по-прежнему 1 января 22:00 в то время как настанет полночь во временной зоне +2
			Time:             time.Date(3333, 1, 2, 0, 0, 0, 0, time.FixedZone("FOO", 2*60*60)),
			ExpectedDateTime: time.Date(3333, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, c := range testCases {
		t.Run(c.Time.String(), func(t *testing.T) {
			dateDTO := adapters.NewEmptyDateDTO(c.Time)
			assert.True(t, dateDTO.Date.Equal(c.ExpectedDateTime), "%s != %s", dateDTO.Date, c.ExpectedDateTime)
		})
	}
}

// вообще глобальное состояние - не лучшая идея, но иногда из правил есть исключения!
// в тестах проще повторно использовать один экземпляр фабрики
var testHourFactory = hour.MustNewFactory(hour.FactoryConfig{
	// 500 недель дают нам достаточно энтропии, чтобы избежать дублирования дат
	// (даже если повторяющиеся даты не проблема)
	MaxWeeksInTheFutureToSet: 500,
	MinUtcHour:               0,
	MaxUtcHour:               24,
})

func newFirebaseRepository(t *testing.T, ctx context.Context) *adapters.FirestoreHourRepository {
	firestoreClient, err := firestore.NewClient(ctx, os.Getenv("GCP_PROJECT"))
	require.NoError(t, err)

	return adapters.NewFirestoreHourRepository(firestoreClient, testHourFactory)
}

func newMySQLRepository(t *testing.T) *adapters.MySQLHourRepository {
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)

	return adapters.NewMySQLHourRepository(db, testHourFactory)
}

func newValidAvailableHour(t *testing.T) *hour.Hour {
	hourTime := newValidHourTime()

	hour, err := testHourFactory.NewAvailableHour(hourTime)
	require.NoError(t, err)

	return hour
}

// usedHours хранит часы, используемые во время теста,
// чтобы гарантировать, что в одном тестовом прогоне мы не используем один и тот же час
// (между тестовыми запусками проблем быть не должно)
var usedHours = sync.Map{}

func newValidHourTime() time.Time {
	for {
		minTime := time.Now().AddDate(0, 0, 1)

		minTimestamp := minTime.Unix()
		maxTimestamp := minTime.AddDate(0, 0, testHourFactory.Config().MaxWeeksInTheFutureToSet*7).Unix()

		t := time.Unix(rand.Int63n(maxTimestamp-minTimestamp)+minTimestamp, 0).Truncate(time.Hour).Local()

		_, alreadyUsed := usedHours.LoadOrStore(t.Unix(), true)
		if !alreadyUsed {
			return t
		}
	}
}

func assertHourInRepository(ctx context.Context, t *testing.T, repo hour.Repository, hour *hour.Hour) {
	require.NotNil(t, hour)

	hourFromRepo, err := repo.GetHour(ctx, hour.Time())
	require.NoError(t, err)

	assert.Equal(t, hour, hourFromRepo)
}
