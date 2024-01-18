package main

import (
	"fmt"
	"testing"

	"github.com/go-chi/chi/v5"
	config "github.com/raindongz/booking-system/internal/configs"
)

func TestRoutes(t *testing.T){
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type){
	case *chi.Mux:
		//test passed
	default: 
	t.Error(fmt.Sprintf("type is not *chi.Mux, is %T", v))
	}
}