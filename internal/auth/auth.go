package auth

import (
	"avitoTask/internal/db"
	errors "avitoTask/internal/errors"
)

func CheckUserCanManageTender(model db.TenderModel, username, organizationId string) errors.HttpError {

	err := model.CheckUserCanManage(username, organizationId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotResponsibleOrganizationError()

	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}
func CheckUserViewTender(model db.TenderModel, username, tenderId string) errors.HttpError {
	err := model.CheckUserView(username, tenderId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotViewTenderError()
	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}

/*
	func (ah AuthHandler) CheckUserCanManageBid(username, autorType, authorId string) error {
		return ah.DbModels.BidModel.CheckUserCanManageBid(username, autorType, authorId)
	}
*/
func CheckUserViewBid(model db.BidModel, username, bidId string) errors.HttpError {
	err := model.CheckUserView(username, bidId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotViewBidError()
	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}

/*
func (ah AuthHandler) CheckUserCanApproveBid(username, tenderId string) error {
	return ah.DbModels.BidModel.CheckUserCanApproveBid(username, tenderId)
}*/
