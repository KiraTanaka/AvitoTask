package auth

import (
	_ "avitoTask/internal/db"

	log "github.com/sirupsen/logrus"
)

/*var db *sqlx.DB

func InitAuth(conn *sqlx.DB) {
	db = conn
}*/

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
	var canView bool
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
	return db.Get(&canView, query, tenderId, username)
}

func CheckUserCanManageBid(username, autorType, authorId string) error {
	var canManage bool
	log.Info("autorType = " + autorType)
	log.Info("authorId = " + authorId)
	log.Info("username = " + username)
	query := `SELECT TRUE
				FROM employee emp
				WHERE emp.username = $1
				AND ('User' = $3 AND emp.id = $2
					OR 'Organization' = $3 AND EXISTS(SELECT 1
														FROM organization_responsible org_r
														WHERE org_r.organization_id = $2 AND org_r.user_id = emp.id))`
	return db.Get(&canManage, query, username, authorId, autorType)
}
func CheckUserViewBid(username, bidId string) error {
	var canView bool
	log.Info("bidId = " + bidId)
	log.Info("username = " + username)
	query := `SELECT true
				FROM bid b
				WHERE b.id = $1
				AND (b.status IN ('Created', 'Canceled') 
						AND (author_type = 'User' AND EXISTS(SELECT 1
															FROM employee emp
															WHERE emp.id = b.author_id AND emp.username = $2)
								OR b.author_type = 'Organization'
									AND EXISTS(SELECT 1
												FROM organization_responsible org_r
													JOIN employee emp ON emp.id = org_r.user_id AND emp.username = $2
												WHERE org_r.organization_id = b.author_id))
					OR
					b.status = 'Published') `
	return db.Get(&canView, query, bidId, username)
}
func CheckUserCanApproveBid(username, tenderId string) error {
	var canManage bool
	log.Info("tenderId = " + tenderId)
	log.Info("username = " + username)
	query := `SELECT TRUE
				FROM tender t
					JOIN organization_responsible org_r ON org_r.organization_id = t.organization_id
					JOIN employee emp ON emp.id = org_r.user_id AND emp.username = $2
				WHERE t.id = $1`
	return db.Get(&canManage, query, tenderId, username)
}
