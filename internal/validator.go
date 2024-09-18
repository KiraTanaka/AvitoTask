package validator

import (
	"avitoTask/internal/db"
)

func CheckUserExists(username string) error {
	return db.CheckUserExists(username)
}
func CheckOrganizationExists(organizationId string) error {
	return db.CheckOrganizationExists(organizationId)
}

func CheckTenderExists(tenderId string) error {
	return db.CheckTenderExists(tenderId)
}

func CheckBidExists(bidId string) error {
	return db.CheckBidExists(bidId)
}
