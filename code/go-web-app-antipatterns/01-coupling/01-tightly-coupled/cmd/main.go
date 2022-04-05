package main

import (
	"github.com/MaksimDzhangirov/three-dots/code/go-web-app-antipatterns/01-tightly-coupled/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root@tcp(mysql)/tightly_coupled?parseTime=true"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	err = db.AutoMigrate(&internal.User{})
	if err != nil {
		log.Fatal("failed to apply migrations")
	}

	err = db.AutoMigrate(&internal.Email{})
	if err != nil {
		log.Fatal("failed to apply migrations")
	}

	storage := internal.NewUserStorage(db)
	h := internal.NewUserHandler(storage)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.GetUsers)
		r.Post("/", h.PostUser)

		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", h.GetUser)
			r.Patch("/", h.PatchUser)
			r.Delete("/", h.DeleteUser)
		})
	})

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
