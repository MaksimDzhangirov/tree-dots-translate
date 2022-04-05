module github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part1/internal/trainer

go 1.16

replace github.com/MaksimDzhangirov/three-dots/internal/common => ../common

require (
	cloud.google.com/go/firestore v1.6.1
	github.com/MaksimDzhangirov/three-dots/internal/common v0.0.0-00010101000000-000000000000
	github.com/deepmap/oapi-codegen v1.9.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/render v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/api v0.59.0
	google.golang.org/grpc v1.40.0
)
