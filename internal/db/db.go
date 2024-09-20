package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	"avitoTask/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type UserModel struct {
	db *sqlx.DB
}

type OrganizationModel struct {
	db *sqlx.DB
}
type DbModels struct {
	UserModel         UserModel
	OrganizationModel OrganizationModel
	TenderModel       TenderModel
	BidModel          BidModel
}

var ErrorNoRows error = sql.ErrNoRows

type Model interface {
	CheckExists(string) error
}

//go:embed queries/checkUserExists.sql
var checkUserExistsQuery string

//go:embed queries/checkOrganizationExists.sql
var checkOrganizationExistsQuery string

func NewDbConnect(config *config.Configuration) (DbModels, error) {
	var dbModels DbModels
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Dbname)

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		log.Error(err)
		return dbModels, err
	}

	err = db.Ping()
	if err != nil {
		log.Error(err)
		return dbModels, err
	}
	log.Info("Connection to the database is completed.")

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Error(err)
		return dbModels, err
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

	dbModels = DbModels{UserModel: UserModel{db: db},
		OrganizationModel: OrganizationModel{db: db},
		TenderModel:       TenderModel{db: db},
		BidModel:          BidModel{db: db}}

	return dbModels, nil
}

func (m UserModel) CheckExists(username string) error {
	var userExists bool
	return m.db.Get(&userExists, checkUserExistsQuery, username)
}
func (m OrganizationModel) CheckExists(organizationId string) error {
	var organizationExists bool
	return m.db.Get(&organizationExists, checkOrganizationExistsQuery, organizationId)
}
