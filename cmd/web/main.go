package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	config "github.com/raindongz/booking-system/internal/configs"
	"github.com/raindongz/booking-system/internal/handlers"
	"github.com/raindongz/booking-system/internal/render"
)

const portNum = ":8080"
var app config.AppConfig
var session *scs.SessionManager

func main(){

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

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("can not create template cache")
	}

	app.TemplateCache = tc
	app.UseCache =false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplate(app)

	// http.HandleFunc("/", handlers.Home)
	// http.HandleFunc("/about", handlers.Repo.About)

	fmt.Println("server is runing on " + portNum)
	// http.ListenAndServe(portNum, nil)

	srv := &http.Server{
		Addr: portNum,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

// func divide(x, y float32) (float32, error){
// 	if y <= 0 {
// 		err := errors.New("can not divide 0")
// 		return 0, err
// 	}

// 	result := x/y
// 	return result, nil
// }

// func addValues(x, y int) (int){
// 	var sum int
// 	sum = x + y
// 	return sum
// }