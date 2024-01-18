package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	config "github.com/raindongz/booking-system/internal/configs"
	mydriver "github.com/raindongz/booking-system/internal/driver"
	"github.com/raindongz/booking-system/internal/handlers"
	"github.com/raindongz/booking-system/internal/helpers"
	"github.com/raindongz/booking-system/internal/models"
	"github.com/raindongz/booking-system/internal/render"
)

const portNum = ":8080"
var app config.AppConfig
var session *scs.SessionManager

var infoLog *log.Logger
var errorLog *log.Logger

func main(){
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	fmt.Println("Starting mail listener")
	listenForMail()


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

func run() (*mydriver.DB,  error) {

	//put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	//read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()
	if *dbName == "" || *dbUser == "" {
		fmt.Println("missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	//change this to true in production
	app.InProduction = *inProduction
	app.UseCache = *useCache

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	//persist after browser window is closed
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
 	//https(local host set http production set to true)
	session.Cookie.Secure = app.InProduction
	app.Session = session

	//connect to database
	log.Println("connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := mydriver.ConnectSQL(connectionString)
	// db, err := mydriver.ConnectSQL("host=localhost port=5432 dbname=bookings user=amdyes password=")
	if err != nil {
		log.Fatal("cannot connect to database.")
	}

	log.Println("connected to database")
	

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("can not create template cache")
		return nil, err
	}

	app.TemplateCache = tc

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	render.NewRenderer(&app)

	//initiallize cntrallized error handler
	helpers.NewHelpers(&app) 

	return db, nil
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