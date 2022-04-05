module github.com/MaksimDzhangirov/three-dots/part15/internal/c4

go 1.16

require (
	github.com/krzysztofreczek/go-structurizr v0.1.5
	github.com/MaksimDzhangirov/three-dots/part15/internal/trainer v0.0.0-00010101000000-000000000000
	github.com/MaksimDzhangirov/three-dots/part15/internal/trainings v0.0.0-00010101000000-000000000000
)

replace (
	github.com/MaksimDzhangirov/three-dots/part15/internal/common => ../../internal/common/
	github.com/MaksimDzhangirov/three-dots/part15/internal/trainer => ../../internal/trainer/
	github.com/MaksimDzhangirov/three-dots/part15/internal/trainings => ../../internal/trainings/
)