package validator

import (
	db "avitoTask/internal/db"
	_ "embed"

	"github.com/jmoiron/sqlx"
)

//go:embed checkUserExists.sql
var checkUserExistsQuery string

func InitValidator(conn *sqlx.DB) {
	db = conn
}

func CheckUserExists(username string) error {
	var userExists bool
	return db.Get(&userExists, checkUserExistsQuery, username)
}
func CheckOrganizationExists(organizationId string) error {
	var organizationExists bool
	return db.Get(&organizationExists, `SELECT TRUE
								FROM   organization
								WHERE  id = $1`, organizationId)
}

func CheckTenderExists(tenderId string) error {
	var tenderExists bool
	return db.Get(&tenderExists, `SELECT TRUE
								FROM   tender
								WHERE  id = $1`, tenderId)
}

func CheckBidExists(bidId string) error {
	var bidExists bool
	return db.Get(&bidExists, `SELECT TRUE
								FROM   bid
								WHERE  id = $1`, bidId)
}
