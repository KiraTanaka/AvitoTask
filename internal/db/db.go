package db

import (
	_ "database/sql"
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

var db *sqlx.DB

//go:embed queries/checkUserExists.sql
var checkUserExistsQuery string

//go:embed queries/checkOrganizationExists.sql
var checkOrganizationExistsQuery string

//go:embed queries/checkTenderExists.sql
var checkTenderExistsQuery string

//go:embed queries/checkBidExists.sql
var checkBidExistsQuery string

//go:embed queries/checkUserCanManageTender.sql
var checkUserCanManageTenderQuery string

//go:embed queries/checkUserViewTender.sql
var checkUserViewTenderQuery string

//go:embed queries/checkUserCanManageBid.sql
var checkUserCanManageBidQuery string

//go:embed queries/checkUserViewBid.sql
var checkUserViewBidQuery string

//go:embed queries/checkUserCanApproveBid.sql
var checkUserCanApproveBidQuery string

func NewDbConnect(config *config.Configuration) (*sqlx.DB, error) {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Dbname)

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

func CheckUserExists(username string) error {
	var userExists bool
	return db.Get(&userExists, checkUserExistsQuery, username)
}
func CheckOrganizationExists(organizationId string) error {
	var organizationExists bool
	return db.Get(&organizationExists, checkOrganizationExistsQuery, organizationId)
}

func CheckTenderExists(tenderId string) error {
	var tenderExists bool
	return db.Get(&tenderExists, checkTenderExistsQuery, tenderId)
}

func CheckBidExists(bidId string) error {
	var bidExists bool
	return db.Get(&bidExists, checkBidExistsQuery, bidId)
}

func CheckUserCanManageTender(username, organizationId string) error {
	var isResponsibleOrganization bool
	return db.Get(&isResponsibleOrganization, checkUserCanManageTenderQuery, organizationId, username)
}
func CheckUserViewTender(username, tenderId string) error {
	var canView bool
	return db.Get(&canView, checkUserViewTenderQuery, tenderId, username)
}

func CheckUserCanManageBid(username, autorType, authorId string) error {
	var canManage bool
	return db.Get(&canManage, checkUserCanManageBidQuery, username, authorId, autorType)
}
func CheckUserViewBid(username, bidId string) error {
	var canView bool
	return db.Get(&canView, checkUserViewBidQuery, bidId, username)
}
func CheckUserCanApproveBid(username, tenderId string) error {
	var canManage bool
	return db.Get(&canManage, checkUserCanApproveBidQuery, tenderId, username)
}
