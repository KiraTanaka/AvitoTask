package auth

import (
	"avitoTask/internal/db"
	errors "avitoTask/internal/errors"
)

func CheckUserCanManageTender(model *db.TenderModel, username, organizationId string) errors.HttpError {

	err := model.CheckUserCanManage(username, organizationId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotResponsibleOrganizationError()

	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}
func CheckUserViewTender(model *db.TenderModel, username, tenderId string) errors.HttpError {
	err := model.CheckUserView(username, tenderId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotViewTenderError()
	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}

func CheckUserCanManageBid(model *db.BidModel, username, autorType, authorId string) errors.HttpError {
	err := model.CheckUserCanManageBid(username, autorType, authorId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotAuthorOrResponsibleOrganizationError()
	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}

func CheckUserViewBid(model *db.BidModel, username, bidId string) errors.HttpError {
	err := model.CheckUserView(username, bidId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotViewBidError()
	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}

func CheckUserCanApproveBid(model *db.BidModel, username, tenderId string) errors.HttpError {
	err := model.CheckUserCanApproveBid(username, tenderId)
	if err == db.ErrorNoRows {
		return errors.GetUserNotResponsibleOrganizationError()
	} else if err != nil {
		return errors.GetInternalServerError(err)

	}
	return errors.HttpError{}
}
