package main

import (
	_ "database/sql"
	"fmt"

	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	"avitoTask/internal/http"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type DbConfig struct {
	/*host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1"
	dbname   = "postgres"*/
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     int    `env:"POSTGRES_PORT" env-required:"true"`
	Dbname   string `env:"POSTGRES_DATABASE" env-required:"true"`
	User     string `env:"POSTGRES_USERNAME" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
}

var db *sqlx.DB

func main() {
	dbConfig, err := ReadDbConfig()
	if err != nil {
		log.Error(err)
		return
	}
	fmt.Println(dbConfig.Host)
	fmt.Println(dbConfig.Port)
	fmt.Println(dbConfig.User)
	fmt.Println(dbConfig.Password)
	fmt.Println(dbConfig.Dbname)
	fmt.Println("new")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Dbname)
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

	routes.Run("avitotask:8080")
}

func ReadDbConfig() (*DbConfig, error) {
	var dbconfig DbConfig
	if err := cleanenv.ReadEnv(&dbconfig); err != nil {
		return nil, fmt.Errorf("DB config error: %w\n", err)
	}
	return &dbconfig, nil
}
