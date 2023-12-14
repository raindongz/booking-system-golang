package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	config "github.com/raindongz/booking-system/pkg/configs"
	"github.com/raindongz/booking-system/pkg/handlers"
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

	return mux
}