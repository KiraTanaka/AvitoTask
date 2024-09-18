package db

import (
	_ "embed"

	"github.com/lib/pq"
)

//go:embed queries/checkTenderExists.sql
var checkTenderExistsQuery string

//go:embed queries/checkUserCanManageTender.sql
var checkUserCanManageTenderQuery string

//go:embed queries/checkUserViewTender.sql
var checkUserViewTenderQuery string

//go:embed queries/getTendersByTypeOfService.sql
var getTendersByTypeOfServiceQuery string

//go:embed queries/getUserTenders.sql
var getUserTendersQuery string

func CheckTenderExists(tenderId string) error {
	var tenderExists bool
	return db.Get(&tenderExists, checkTenderExistsQuery, tenderId)
}

func CheckUserCanManageTender(username, organizationId string) error {
	var isResponsibleOrganization bool
	return db.Get(&isResponsibleOrganization, checkUserCanManageTenderQuery, organizationId, username)
}
func CheckUserViewTender(username, tenderId string) error {
	var canView bool
	return db.Get(&canView, checkUserViewTenderQuery, tenderId, username)
}
func GetTendersByTypeOfService(tenders interface{}, serviceTypes []string, limit string, offset string) error {
	return db.Select(&tenders, getTendersByTypeOfServiceQuery, pq.Array(serviceTypes), len(serviceTypes), limit, offset)
}

func GetUserTenders(tenders interface{}, username, limit, offset string) error {
	return db.Select(&tenders, getUserTendersQuery, username, limit, offset)
}
