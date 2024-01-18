package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	config "github.com/raindongz/booking-system/internal/configs"
	"github.com/raindongz/booking-system/internal/models"
	"github.com/raindongz/booking-system/internal/render"
)

var functions = template.FuncMap{}
var app config.AppConfig
var session *scs.SessionManager
var pathToTemplate = "./../../templates"

func getRoutes() http.Handler{
	//put in the session
	gob.Register(models.Reservation{})

	//change this to true in production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	//persist after browser window is closed
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
 	//https(local host set http production set to true)
	session.Cookie.Secure = app.InProduction
	app.Session = session


	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("can not create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = true 

	repo := NewRepo(&app)
	NewHandlers(repo)

	render.NewRenderer(&app)


	mux := chi.NewRouter()

	//middleware: recover
	mux.Use(middleware.Recoverer)
	//mux.Use(WriteToConsole)

	// mux.Use(NoSurf)

	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/general", Repo.General)
	mux.Get("/major", Repo.Major)
	mux.Get("/book", Repo.Book)
	mux.Post("/book", Repo.PostBook)
	mux.Post("/book-json", Repo.PostBookJson)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	mux.Get("/contact", Repo.Contact)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux

}

//nosurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler{
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

//sessionload loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler{
	//this session is from main package
	return session.LoadAndSave(next)
}

func CreateTestTemplateCache()(map[string]*template.Template, error){
	myCache := map[string]*template.Template{}
	//get all the files named *.page.tmpl from ./template
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplate))
	if err != nil {
		return myCache, err
	}

	//range through all the files ending with 
	for _, page := range pages{
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
		if err !=nil {
			return myCache, err
		}

		if len(matches) > 0{
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
			if err !=nil {
				return myCache, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}