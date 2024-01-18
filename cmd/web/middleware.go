package main

import (
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/raindongz/booking-system/internal/helpers"
)

// func WriteToConsole(next http.Handler) http.Handler{
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
// 		fmt.Println("Hit the page")
// 		next.ServeHTTP(w, r)
// 	})
// }

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

func Auth(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if !helpers.IsAuthenticate(r){
			session.Put(r.Context(), "error", "login first")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}