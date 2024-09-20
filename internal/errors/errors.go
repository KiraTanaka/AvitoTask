package errors

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type HttpError struct {
	httpCode int
	Reason   string `json:"reason"`
}

func (e HttpError) SeparateCode() (int, HttpError) {
	return e.httpCode, e
}
func (e HttpError) IsEmpty() bool {
	return e.Reason == ""
}

var (
	//400
	OrganizationNotExistsOrIncorrectError = HttpError{http.StatusBadRequest, "Организация не существует или некорректна."}
	NewStatusNotPassedError               = HttpError{http.StatusBadRequest, "Новый статус должен быть указан."}
	TenderIdNotPassedError                = HttpError{http.StatusBadRequest, "Идентификатор тендера должен быть указан."}
	BidIdNotPassedError                   = HttpError{http.StatusBadRequest, "Идентификатор предложения должен быть указан."}
	AuthorNotFoundError                   = HttpError{http.StatusBadRequest, "Указанный автор не существует."}
	InvalidVersionError                   = HttpError{http.StatusBadRequest, "Указанная версия больше или равна текущей версии тендера."}
	InvalidServiceTypeError               = HttpError{http.StatusBadRequest, "Недопустимый вид услуги"}
	InvalidStatusError                    = HttpError{http.StatusBadRequest, "Недопустимый статус"}
	InvalidDecisionError                  = HttpError{http.StatusBadRequest, "Недопустимое решение"}
	DecisionNotPassedError                = HttpError{http.StatusBadRequest, "Решение должено быть указано."}
	BidAlreadyHasDecisionError            = HttpError{http.StatusBadRequest, "Решение по предложению уже принято."}
	UserHasDecisionForBidError            = HttpError{http.StatusBadRequest, "Вы уже приняли решение по предложению."}
	// 401
	UserNotPassedError            = HttpError{http.StatusUnauthorized, "Пользователь должен быть указан."}
	UserNotExistsOrIncorrectError = HttpError{http.StatusUnauthorized, "Пользователь не существует или некорректен."}
	//403
	UserNotResponsibleOrganizationError         = HttpError{http.StatusForbidden, "Необходимо быть ответственным за организацию."}
	UserNotAuthorOrResponsibleOrganizationError = HttpError{http.StatusForbidden, "Необходимо быть автором или ответственным за организацию."}
	UserNotViewTenderError                      = HttpError{http.StatusForbidden, "Нельзя просматривать неопубликованные тендеры, если вы не ответственный за организацию."}
	UserNotViewBidError                         = HttpError{http.StatusForbidden, "Нельзя просматривать неопубликованные предложения, если вы не ответственный за организацию или автор."}
	//404
	TenderNotFoundError  = HttpError{http.StatusNotFound, "Указанный тендер не существует."}
	BidNotFoundError     = HttpError{http.StatusNotFound, "Указанное предложение не существует."}
	VersionNotFoundError = HttpError{http.StatusNotFound, "Версия не найдена."}
)

// 400 (StatusBadRequest) - Данные неправильно сформированы или не соответствуют требованиям.

func GetInvalidRequestFormatOrParametersError(err error) HttpError {
	log.Error(err)
	return HttpError{http.StatusBadRequest, err.Error()}
}

func GetNewStatusNotPassedError() HttpError {
	log.Error(NewStatusNotPassedError)
	return NewStatusNotPassedError
}
func GetTenderIdNotPassedError() HttpError {
	log.Error(TenderIdNotPassedError)
	return TenderIdNotPassedError
}

func GetBidIdNotPassedError() HttpError {
	log.Error(BidIdNotPassedError)
	return BidIdNotPassedError
}
func GetAuthorNotFoundError() HttpError {
	log.Error(AuthorNotFoundError)
	return AuthorNotFoundError
}

func GetOrganizationNotExistsOrIncorrectError() HttpError {
	log.Error(OrganizationNotExistsOrIncorrectError)
	return OrganizationNotExistsOrIncorrectError
}

func GetInvalidVersionError() HttpError {
	log.Error(InvalidVersionError)
	return InvalidVersionError
}

func GetInvalidServiceTypeError() HttpError {
	log.Error(InvalidServiceTypeError)
	return InvalidServiceTypeError
}

func GetInvalidStatusError() HttpError {
	log.Error(InvalidStatusError)
	return InvalidStatusError
}
func GetDecisionNotPassedError() HttpError {
	log.Error(DecisionNotPassedError)
	return DecisionNotPassedError
}
func GetInvalidDecisionError() HttpError {
	log.Error(InvalidDecisionError)
	return InvalidDecisionError
}

func GetBidAlreadyHasDecisionError() HttpError {
	log.Error(BidAlreadyHasDecisionError)
	return BidAlreadyHasDecisionError
}

func GetUserHasDecisionForBidError() HttpError {
	log.Error(UserHasDecisionForBidError)
	return UserHasDecisionForBidError
}

// 401 (StatusUnauthorized) - Пользователь не существует или некорректен.

func GetUserNotPassedError() HttpError {
	log.Error(UserNotPassedError)
	return UserNotPassedError
}

func GetUserNotExistsOrIncorrectError() HttpError {
	log.Error(UserNotExistsOrIncorrectError)
	return UserNotExistsOrIncorrectError
}

// 403 (StatusForbidden) - Недостаточно прав для выполнения действия.

func GetUserNotResponsibleOrganizationError() HttpError {
	log.Error(UserNotResponsibleOrganizationError)
	return UserNotResponsibleOrganizationError
}
func GetUserNotAuthorOrResponsibleOrganizationError() HttpError {
	log.Error(UserNotAuthorOrResponsibleOrganizationError)
	return UserNotAuthorOrResponsibleOrganizationError
}

func GetUserNotViewTenderError() HttpError {
	log.Error(UserNotViewTenderError)
	return UserNotViewTenderError
}
func GetUserNotViewBidError() HttpError {
	log.Error(UserNotViewBidError)
	return UserNotViewBidError
}

// 404 (StatusNotFound) - Тендер или предложение не найдено.

func GetTenderNotFoundError() HttpError {
	log.Error(TenderNotFoundError.Reason)
	return TenderNotFoundError
}
func GetVersionNotFoundError() HttpError {
	log.Error(VersionNotFoundError)
	return VersionNotFoundError
}
func GetBidNotFoundError() HttpError {
	log.Error(BidNotFoundError)
	//c.AbortWithStatusJSON(http.StatusNotFound, BidNotFoundError)
	return BidNotFoundError
}

// 500 (StatusInternalServerError) - Сервер не готов обрабатывать запросы, если ответ статусом 500 или любой другой, кроме 200.

func GetInternalServerError(err error) HttpError {
	log.Error(err)
	return HttpError{http.StatusInternalServerError, err.Error()}
}
