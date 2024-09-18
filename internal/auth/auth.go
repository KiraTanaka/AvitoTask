package auth

import (
	"avitoTask/internal/db"
)

func CheckUserCanManageTender(username, organizationId string) error {
	return db.CheckUserCanManageTender(username, organizationId)
}
func CheckUserViewTender(username, tenderId string) error {
	return db.CheckUserViewTender(username, tenderId)
}

func CheckUserCanManageBid(username, autorType, authorId string) error {
	return db.CheckUserCanManageBid(username, autorType, authorId)
}
func CheckUserViewBid(username, bidId string) error {
	return db.CheckUserViewBid(username, bidId)
}
func CheckUserCanApproveBid(username, tenderId string) error {
	return db.CheckUserCanApproveBid(username, tenderId)
}
