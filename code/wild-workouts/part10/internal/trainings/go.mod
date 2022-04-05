module github.com/MaksimDzhangirov/three-dots/part10/internal/trainings

go 1.16

replace github.com/MaksimDzhangirov/three-dots/part10/internal/common => ../common

require (
	cloud.google.com/go v0.38.0
	github.com/MaksimDzhangirov/three-dots/part10/internal/common v0.0.0-00010101000000-000000000000
	github.com/deepmap/oapi-codegen v1.9.0
	github.com/go-chi/chi/v5 v5.0.7
	github.com/go-chi/render v1.0.1
	github.com/google/uuid v1.3.0
	google.golang.org/api v0.21.0
)
