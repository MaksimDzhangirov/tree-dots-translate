module github.com/MaksimDzhangirov/three-dots/part10/internal/trainer

go 1.16

replace github.com/MaksimDzhangirov/three-dots/part10/internal/common => ../common

require (
	cloud.google.com/go/firestore v1.6.1
	github.com/MaksimDzhangirov/three-dots/part10/internal/common v0.0.0-00010101000000-000000000000
	github.com/deepmap/oapi-codegen v1.9.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/render v1.0.1
	github.com/golang/protobuf v1.5.2
	github.com/jmoiron/sqlx v1.3.4
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/multierr v1.1.0
	google.golang.org/api v0.59.0
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
)
