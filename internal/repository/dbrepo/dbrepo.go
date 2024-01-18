package dbrepo

import (
	"database/sql"

	config "github.com/raindongz/booking-system/internal/configs"
	"github.com/raindongz/booking-system/internal/repository"
)

type postgresDBRepo struct{
	App *config.AppConfig
	DB *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo{
	return &postgresDBRepo{
		App: a,
		DB: conn,
	}
}