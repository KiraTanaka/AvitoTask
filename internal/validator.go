package validator

import (
	"slices"
	"strconv"

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

func CheckBid(model db.Model, bidId string) errors.HttpError {
	if bidId == "" {
		return errors.GetBidIdNotPassedError()
	}
	err := uuid.Validate(bidId)
	if err != nil {
		return errors.GetInvalidRequestFormatOrParametersError(err)
	}
	err = model.CheckExists(bidId)
	if err == db.ErrorNoRows {
		return errors.GetBidNotFoundError()
	} else if err != nil {
		return errors.GetInternalServerError(err)
	}
	return errors.HttpError{}
}

func ServiceTypeIsAcceptable(serviceType string, serviceTypesConst []string) errors.HttpError {
	if !slices.Contains(serviceTypesConst, serviceType) {
		return errors.GetInvalidServiceTypeError()
	}
	return errors.HttpError{}
}

func CheckServiceTypes(serviceTypes []string, serviceTypesConst []string) errors.HttpError {
	for _, serviceType := range serviceTypes {
		errHttp := ServiceTypeIsAcceptable(serviceType, serviceTypesConst)
		if !errHttp.IsEmpty() {
			return errHttp
		}
	}
	return errors.HttpError{}
}

func AuthorTypeAcceptable(authorType string, authorTypesConst []string) errors.HttpError {
	if !slices.Contains(authorTypesConst, authorType) {
		return errors.GetInvalidAuthorTypeError()
	}
	return errors.HttpError{}
}

func CheckStatus(status string, statusesConst []string) errors.HttpError {
	if status == "" {
		return errors.GetNewStatusNotPassedError()
	} else if !slices.Contains(statusesConst, status) {
		return errors.GetInvalidStatusError()
	}
	return errors.HttpError{}
}

func CheckVersion(param string) (int, errors.HttpError) {
	version, err := strconv.Atoi(param)
	if err != nil {
		return 0, errors.GetInvalidVersionError()
	}
	return version, errors.HttpError{}
}

func CheckBidDecision(decision string, decisionTypesConst []string) errors.HttpError {
	if decision == "" {
		return errors.GetDecisionNotPassedError()
	} else if !slices.Contains(decisionTypesConst, decision) {
		return errors.GetInvalidDecisionError()
	}
	return errors.HttpError{}
}
