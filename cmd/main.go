package main

import (
	_ "database/sql"
	"fmt"

	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	"avitoTask/internal/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1"
	dbname   = "postgres"
)

var db *sqlx.DB

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	auth.InitAuth(db)
	validator.InitValidator(db)
	routes := http.InitRoutes(db)

	routes.Run("localhost:8080")
}
