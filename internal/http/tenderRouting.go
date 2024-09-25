package http

import (
	"encoding/json"
	"net/http"

	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	db "avitoTask/internal/db"
	"avitoTask/internal/errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type TenderHandler struct {
	tender       *db.TenderModel
	user         *db.UserModel
	organization *db.OrganizationModel
}

type TenderDto struct {
	Id          string `json:"id" binding:"max=100"`
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"required,max=500"`
	ServiceType string `json:"serviceType" binding:"required,oneof=Construction Delivery Manufacture"`
	Status      string `json:"status" binding:"required,oneof=Created Published Closed"`
	Version     int    `json:"version" binding:"required,min=1"`
	CreatedAt   string `json:"createdAt" binding:"required"`
}

var statusesConst []string = []string{"Created", "Published", "Closed"}
var serviceTypesConst []string = []string{"Construction", "Delivery", "Manufacture"}

func InitTenderRoutes(routes *gin.RouterGroup, tenderHandler *TenderHandler) {
	tenderRoutes := routes.Group("/tenders")
	//GET
	tenderRoutes.GET("/", tenderHandler.getTenders)
	tenderRoutes.GET("/my", tenderHandler.getUserTender)
	tenderRoutes.GET("/:tenderId/status", tenderHandler.getStatusTender)
	//POST
	tenderRoutes.POST("/new", tenderHandler.createTender)
	//PUT
	tenderRoutes.PUT("/:tenderId/status", tenderHandler.changeStatusTender)
	tenderRoutes.PUT("/:tenderId/rollback/:version", tenderHandler.rollbackVersionTender)
	//PATCH
	tenderRoutes.PATCH("/:tenderId/edit", tenderHandler.editTender)

}

func convertToDto(t *db.Tender) *TenderDto {
	return &TenderDto{
		Id:          t.Id,
		Name:        t.Name,
		Description: t.Description,
		ServiceType: t.ServiceType,
		Status:      t.Status,
		Version:     t.Version,
		CreatedAt:   t.CreatedAt,
	}
}

func (h *TenderHandler) getTenders(c *gin.Context) {
	log.Info("Чтение параметров")
	limit, offset := SetDefaultPaginationParamIfEmpty(c.Query("limit"), c.Query("offset"))

	log.Info("Валидация")
	serviceTypes := c.QueryArray("service_type")
	errHttp := validator.CheckServiceTypes(serviceTypes, serviceTypesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных")
	tenders, err := h.tender.GetListByTypeOfService(serviceTypes, limit, offset)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}
	var tendersDto []TenderDto
	for _, tender := range *tenders {
		tendersDto = append(tendersDto, *convertToDto(&tender))
	}

	c.JSON(http.StatusOK, tendersDto)
}

func (h *TenderHandler) getUserTender(c *gin.Context) {
	log.Info("Чтение параметров")
	limit, offset := SetDefaultPaginationParamIfEmpty(c.Query("limit"), c.Query("offset"))
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckUser(h.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных")
	tenders, err := h.tender.GetListForUser(username, limit, offset)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	tendersDto := []TenderDto{}
	for _, tender := range *tenders {
		tendersDto = append(tendersDto, *convertToDto(&tender))
	}

	c.JSON(http.StatusOK, tendersDto)
}

func (h *TenderHandler) getStatusTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckUser(h.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckTender(h.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserViewTender(h.tender, username, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных")
	status, err := h.tender.GetStatus(tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *TenderHandler) createTender(c *gin.Context) {
	log.Info("Чтение параметров")
	someTender := db.TenderDefault()
	err := c.BindJSON(someTender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInvalidRequestFormatOrParametersError(err).SeparateCode())
		return
	}

	log.Info("Валидация")
	errHttp := validator.ServiceTypeIsAcceptable(someTender.ServiceType, serviceTypesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckUser(h.user, someTender.CreatorUsername)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckOrganization(h.organization, someTender.OrganizationId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageTender(h.tender, someTender.CreatorUsername, someTender.OrganizationId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Создание")
	err = h.tender.Create(someTender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(someTender))
}

func (h *TenderHandler) changeStatusTender(c *gin.Context) {
	log.Info("Чтение параметров")
	status := c.Query("status")
	username := c.Query("username")
	tenderId := c.Param("tenderId")

	log.Info("Валидация")
	errHttp := validator.CheckStatus(status, statusesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckTender(h.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckUser(h.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение дополнительных данных")
	tender := db.TenderDefault()
	tender, err := h.tender.Get(tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageTender(h.tender, username, tender.OrganizationId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Изменение")
	err = h.tender.ChangeStatus(status, tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение измененных данных")
	tender, err = h.tender.Get(tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(tender))
}

func (h *TenderHandler) editTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckTender(h.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckUser(h.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение исходных данных")
	tender, err := h.tender.Get(tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение новых значений")
	err = c.BindJSON(tender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInvalidRequestFormatOrParametersError(err).SeparateCode())
		return
	}

	errHttp = validator.ServiceTypeIsAcceptable(tender.ServiceType, serviceTypesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageTender(h.tender, username, tender.OrganizationId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Изменение")
	err = h.tender.Edit(tender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение измененных данных")
	tender, err = h.tender.Get(tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(tender))
}
func (h *TenderHandler) rollbackVersionTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckTender(h.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	version, errHttp := validator.CheckVersion(c.Param("version"))
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckUser(h.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение исходных данных")
	tender, err := h.tender.Get(tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	if version >= tender.Version {
		c.AbortWithStatusJSON(errors.GetVersionIsOutOfBoundsError().SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageTender(h.tender, username, tender.OrganizationId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных версии")
	params, err := h.tender.GetParamsByVersion(tender.Id, version)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	err = json.Unmarshal([]byte(params), &tender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Откат до версии")
	err = h.tender.Edit(tender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение измененных данных")
	tender, err = h.tender.Get(tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(tender))
}
