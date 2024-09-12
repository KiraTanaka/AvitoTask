package error

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type InternalErrorBody struct {
	Reason string `json:"reason"`
}

var (
	UserNotPassedError                    = InternalErrorBody{"Пользователь должен быть указан."}
	UserNotExistsOrIncorrectError         = InternalErrorBody{"Пользователь не существует или некорректен."}
	OrganizationNotExistsOrIncorrectError = InternalErrorBody{"Организация не существует или некорректна."}
	NewStatusNotPassedError               = InternalErrorBody{"Новый статус должен быть указан."}
	TenderIdNotPassedError                = InternalErrorBody{"Идентификатор тендера должен быть указан."}
	TenderNotFoundError                   = InternalErrorBody{"Указанный тендер не существует."}
	BidIdNotPassedError                   = InternalErrorBody{"Идентификатор предложения должен быть указан."}
	AuthorNotFoundError                   = InternalErrorBody{"Указанный автор не существует."}
	UserNotResponsibleOrganizationError   = InternalErrorBody{"Необходимо быть ответственным за организацию."}
	InvalidVersionError                   = InternalErrorBody{"Указанная версия больше или равна текущей версии тендера."}
	VersionNotFoundError                  = InternalErrorBody{"Версия не найдена."}
)

// 400 (StatusBadRequest) - Данные неправильно сформированы или не соответствуют требованиям.

func GetInvalidRequestFormatOrParametersError(c *gin.Context, err error) {
	log.Error(err)
	c.AbortWithStatusJSON(http.StatusBadRequest, InternalErrorBody{err.Error()})
}

func GetNewStatusNotPassedError(c *gin.Context) {
	log.Error(NewStatusNotPassedError)
	c.AbortWithStatusJSON(http.StatusBadRequest, NewStatusNotPassedError)
}

func GetBidIdNotPassedError(c *gin.Context) {
	log.Error(BidIdNotPassedError)
	c.AbortWithStatusJSON(http.StatusBadRequest, BidIdNotPassedError)
}
func GetAuthorNotFoundError(c *gin.Context) {
	log.Error(AuthorNotFoundError)
	c.AbortWithStatusJSON(http.StatusBadRequest, AuthorNotFoundError)
}

func GetOrganizationNotExistsOrIncorrectError(c *gin.Context) {
	log.Error(OrganizationNotExistsOrIncorrectError)
	c.AbortWithStatusJSON(http.StatusBadRequest, OrganizationNotExistsOrIncorrectError)
}

func GetInvalidVersionError(c *gin.Context) {
	log.Error(InvalidVersionError)
	c.AbortWithStatusJSON(http.StatusBadRequest, InvalidVersionError)
}

// 401 (StatusUnauthorized) - Пользователь не существует или некорректен.

func GetUserNotPassedError(c *gin.Context) {
	log.Error(UserNotPassedError)
	c.AbortWithStatusJSON(http.StatusUnauthorized, UserNotPassedError)
}

func GetUserNotExistsOrIncorrectError(c *gin.Context) {
	log.Error(UserNotExistsOrIncorrectError)
	c.AbortWithStatusJSON(http.StatusUnauthorized, UserNotExistsOrIncorrectError)
}

// 403 (StatusForbidden) - Недостаточно прав для выполнения действия.

func GetUserNotResponsibleOrganizationError(c *gin.Context) {
	log.Error(UserNotResponsibleOrganizationError)
	c.AbortWithStatusJSON(http.StatusForbidden, UserNotResponsibleOrganizationError)
}

// 404 (StatusNotFound) - Тендер или предложение не найдено.

func GetTenderNotFoundError(c *gin.Context) {
	log.Error(TenderNotFoundError)
	c.AbortWithStatusJSON(http.StatusNotFound, TenderNotFoundError)
}
func GetVersionNotFoundError(c *gin.Context) {
	log.Error(VersionNotFoundError)
	c.AbortWithStatusJSON(http.StatusNotFound, VersionNotFoundError)
}

// 500 (StatusInternalServerError) - Сервер не готов обрабатывать запросы, если ответ статусом 500 или любой другой, кроме 200.

func GetInternalServerError(c *gin.Context, err error) {
	log.Error(err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, InternalErrorBody{err.Error()})
}
