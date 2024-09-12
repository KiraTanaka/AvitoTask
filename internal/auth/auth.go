package auth

import (
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func InitAuth(conn *sqlx.DB) {
	db = conn
}

func CheckUserIsResponsibleOrganization(username, organizationId string) (bool, error) {
	var isResponsibleOrganization bool
	query := `SELECT true
				FROM organization_responsible org_r
					JOIN employee emp ON emp.id = org_r.user_id
				WHERE org_r.organization_id = $1 AND emp.username = $2`
	err := db.Select(&isResponsibleOrganization, query, organizationId, username)
	return isResponsibleOrganization, err
}
