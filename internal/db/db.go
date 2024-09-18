package db

import (
	_ "database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type DbConfig struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     int    `env:"POSTGRES_PORT" env-required:"true"`
	Dbname   string `env:"POSTGRES_DATABASE" env-required:"true"`
	User     string `env:"POSTGRES_USERNAME" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
}

var db *sqlx.DB

func NewDbConnect() (*sqlx.DB, error) {
	dbConfig, err := ReadDbConfig()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Dbname)

	db, err = sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Info("Connection to the database is completed.")

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Error(err)
	}
	err = m.Up()
	if err != nil {
		log.Error(err)
	}
	log.Info("Verification and application of missing migrations is completed.")
	return db, nil
}

func ReadDbConfig() (*DbConfig, error) {
	var dbconfig DbConfig
	err := cleanenv.ReadConfig(".env", &dbconfig)
	//err := cleanenv.ReadEnv(&dbconfig)
	if err != nil {
		return nil, fmt.Errorf("DB config error: %w", err)
	}

	log.Info("Reading of the database configuration parameters is completed.")
	return &dbconfig, nil
}
