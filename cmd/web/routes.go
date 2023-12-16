package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	config "github.com/raindongz/booking-system/internal/configs"
	"github.com/raindongz/booking-system/internal/handlers"
)

func routes(app *config.AppConfig) http.Handler{
	mux := chi.NewRouter()

	//middleware: recover
	mux.Use(middleware.Recoverer)
	//mux.Use(WriteToConsole)

	mux.Use(NoSurf)

	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/general", handlers.Repo.General)
	mux.Get("/major", handlers.Repo.Major)
	mux.Get("/book", handlers.Repo.Book)
	mux.Post("/book", handlers.Repo.PostBook)
	mux.Post("/book-json", handlers.Repo.PostBookJson)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Get("/contact", handlers.Repo.Contact)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}