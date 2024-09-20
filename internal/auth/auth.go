package auth

import (
	"avitoTask/internal/db"
)

type AuthHandler struct {
	DbModels db.DbModels
}

func (ah AuthHandler) CheckUserCanManageTender(username, organizationId string) error {
	return ah.DbModels.TenderModel.CheckUserCanManageTender(username, organizationId)
}
func (ah AuthHandler) CheckUserViewTender(username, tenderId string) error {
	return ah.DbModels.TenderModel.CheckUserViewTender(username, tenderId)
}

func (ah AuthHandler) CheckUserCanManageBid(username, autorType, authorId string) error {
	return ah.DbModels.BidModel.CheckUserCanManageBid(username, autorType, authorId)
}
func (ah AuthHandler) CheckUserViewBid(username, bidId string) error {
	return ah.DbModels.BidModel.CheckUserViewBid(username, bidId)
}
func (ah AuthHandler) CheckUserCanApproveBid(username, tenderId string) error {
	return ah.DbModels.BidModel.CheckUserCanApproveBid(username, tenderId)
}
