module github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part8/internal/users

go 1.16

replace github.com/MaksimDzhangirov/three-dots/internal/common => ../common

require (
	cloud.google.com/go/firestore v1.6.1
	github.com/MaksimDzhangirov/three-dots/internal/common v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.0.7
)
