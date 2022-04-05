package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/MaksimDzhangirov/three-dots/part12/internal/common/client"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainer/app"
	"github.com/MaksimDzhangirov/three-dots/part12/internal/trainer/app/query"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const daysToSet = 30

func loadFixtures(app app.Application) {
	start := time.Now()
	ctx := context.Background()

	logrus.Debug("Waiting for trainer service")
	working := client.WaitForTrainerService(time.Second * 30)
	if !working {
		logrus.Error("Trainer gRPC service is not up")
		return
	}

	logrus.WithField("after", time.Now().Sub(start)).Debug("Trainer service is available")

	if !canLoadFixtures(app, ctx) {
		logrus.Debug("Trainer fixtures are already loaded")
		return
	}

	for {
		err := loadTrainerFixtures(ctx, app)
		if err == nil {
			break
		}

		logrus.WithError(err).Error("Cannot load trainer fixtures")
		time.Sleep(10 * time.Second)
	}

	logrus.WithField("after", time.Now().Sub(start)).Debug("Trainer fixtures loaded")
}

func loadTrainerFixtures(ctx context.Context, application app.Application) error {
	maxDate := time.Now().Add(time.Hour * 24 * daysToSet)
	localRand := rand.New(rand.NewSource(3))

	for date := time.Now(); date.Before(maxDate); date = date.AddDate(0, 0, 1) {
		for hour := 12; hour <= 20; hour++ {
			trainingTime := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, time.UTC)

			if trainingTime.Add(time.Hour).Before(time.Now()) {
				// этот час уже начался
				continue
			}

			if localRand.NormFloat64() > 0 {
				err := application.Commands.MakeHoursAvailable.Handle(ctx, []time.Time{trainingTime})
				if err != nil {
					return errors.Wrap(err, "unable to update hour")
				}
			}
		}
	}

	return nil
}

func canLoadFixtures(app app.Application, ctx context.Context) bool {
	for {
		dates, err := app.Queries.TrainerAvailableHours.Handle(ctx, query.AvailableHours{
			From: time.Now(),
			To:   time.Now().AddDate(0, 0, daysToSet),
		})
		if err == nil {
			for _, date := range dates {
				for _, hour := range date.Hours {
					if hour.Available {
						// нам не нужны фикстуры, если час уже доступен для тренировки
						return false
					}
				}
			}

			return true
		}

		logrus.WithError(err).Error("Cannot check if fixtures can be loaded")
		time.Sleep(10 * time.Second)
	}
}
