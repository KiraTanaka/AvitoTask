package validator

import (
	"avitoTask/internal/db"
	errors "avitoTask/internal/errors"

	"github.com/google/uuid"
)

func CheckUser(model db.Model, username string) errors.HttpError {
	if username == "" {
		return errors.GetUserNotPassedError()
	}

	err := model.CheckExists(username)
	if err == db.ErrorNoRows {
		return errors.GetUserNotExistsOrIncorrectError()
	} else if err != nil {
		return errors.GetInternalServerError(err)
	}
	return errors.HttpError{}
}

func CheckOrganization(model db.Model, organizationId string) errors.HttpError {

	err := uuid.Validate(organizationId)
	if err != nil {
		return errors.GetInvalidRequestFormatOrParametersError(err)
	}

	err = model.CheckExists(organizationId)
	if err == db.ErrorNoRows {
		return errors.GetOrganizationNotExistsOrIncorrectError()
	} else if err != nil {
		return errors.GetInternalServerError(err)
	}
	return errors.HttpError{}
}

func CheckTender(model db.Model, tenderId string) errors.HttpError {
	if tenderId == "" {
		return errors.GetTenderIdNotPassedError()
	}
	err := uuid.Validate(tenderId)
	if err != nil {
		return errors.GetInvalidRequestFormatOrParametersError(err)
	}
	err = model.CheckExists(tenderId)
	if err == db.ErrorNoRows {
		return errors.GetTenderNotFoundError()
	} else if err != nil {

		return errors.GetInternalServerError(err)
	}
	return errors.HttpError{}
}

/*
func CheckBidExists(bidId string) error {
	return vh.DbModels.BidModel.CheckBidExists(bidId)
}*/
