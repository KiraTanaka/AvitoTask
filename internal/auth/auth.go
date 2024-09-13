package auth

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

var db *sqlx.DB

func InitAuth(conn *sqlx.DB) {
	db = conn
}

func CheckUserCanManageTender(username, organizationId string) error {
	var isResponsibleOrganization bool
	log.Info("organizationId = " + organizationId)
	log.Info("username = " + username)
	query := `SELECT true
				FROM organization_responsible org_r
					JOIN employee emp ON emp.id = org_r.user_id
				WHERE org_r.organization_id = $1 AND emp.username = $2`
	return db.Get(&isResponsibleOrganization, query, organizationId, username)
}
func CheckUserViewTender(username, tenderId string) error {
	var isResponsibleOrganization bool
	log.Info("tenderId = " + tenderId)
	log.Info("username = " + username)
	query := `SELECT true
				FROM   tender t
				WHERE  id = $1
					AND ( t.status = 'Published'
							OR EXISTS(SELECT 1
										FROM   organization_responsible org_r
											join employee emp
												ON emp.id = org_r.user_id
													AND emp.username = $2
										WHERE  org_r.organization_id = t.organization_id)
								AND t.status IN ( 'Created', 'Closed' ) ) `
	return db.Get(&isResponsibleOrganization, query, tenderId, username)
}
