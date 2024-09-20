package http

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	validator "avitoTask/internal"
	db "avitoTask/internal/db"
	"avitoTask/internal/errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type TenderHandler struct {
	tender       db.TenderModel
	user         db.UserModel
	organization db.OrganizationModel
}

/*
type tender struct {
	Id              string `json:"id" db:"id" binding:"max=100"`
	Name            string `json:"name" db:"name" binding:"required,max=100"`
	Description     string `json:"description" db:"description" binding:"required,max=500"`
	ServiceType     string `json:"serviceType" db:"service_type" binding:"required,oneof=Construction Delivery Manufacture"`
	Status          string `json:"status" db:"status" binding:"required,oneof=Created Published Closed"`
	Version         int    `json:"version" db:"version" binding:"required,min=1"`
	OrganizationId  string `json:"organizationId" db:"organization_id" binding:"required,max=100"`
	CreatedAt       string `json:"createdAt" db:"created_at" binding:"required"`
	CreatorUsername string `json:"creatorUsername"`
}*/

type tenderDto struct {
	Id          string `json:"id" binding:"max=100"`
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"required,max=500"`
	ServiceType string `json:"serviceType" binding:"required,oneof=Construction Delivery Manufacture"`
	Status      string `json:"status" binding:"required,oneof=Created Published Closed"`
	Version     int    `json:"version" binding:"required,min=1"`
	CreatedAt   string `json:"createdAt" binding:"required"`
}

var StatusConst []string = []string{"Created", "Published", "Closed"}
var ServiceTypesConst []string = []string{"Construction", "Delivery", "Manufacture"}

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

func convertToDto(t db.Tender) tenderDto {
	var tenderDto tenderDto
	tenderDto.Id = t.Id
	tenderDto.Name = t.Name
	tenderDto.Description = t.Description
	tenderDto.ServiceType = t.ServiceType
	tenderDto.Status = t.Status
	tenderDto.Version = t.Version
	tenderDto.CreatedAt = t.CreatedAt
	return tenderDto
}

func (th TenderHandler) getTenders(c *gin.Context) {
	log.Info("Чтение параметров")
	limit := c.Query("limit")
	offset := c.Query("offset")

	log.Info("Валидация")
	if limit == "" {
		limit = "5"
	}
	if offset == "" {
		offset = "0"
	}
	serviceTypes := c.QueryArray("service_type")
	for _, serviceType := range serviceTypes {
		if !slices.Contains(ServiceTypesConst, serviceType) {
			c.AbortWithStatusJSON(errors.GetInvalidServiceTypeError().SeparateCode())
			return
		}
	}

	log.Info("Чтение данных")
	tenders, err := th.tender.GetListByTypeOfService(serviceTypes, limit, offset)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}
	tendersDto := []tenderDto{}
	for _, tender := range tenders {
		tendersDto = append(tendersDto, convertToDto(tender))
	}

	c.JSON(http.StatusOK, tendersDto)
}

func (th TenderHandler) getUserTender(c *gin.Context) {
	log.Info("Чтение параметров")
	limit := c.Query("limit")
	offset := c.Query("offset")
	username := c.Query("username")
	log.Info("Валидация")
	if limit == "" {
		limit = "5"
	}
	if offset == "" {
		offset = "0"
	}
	errHttp := validator.CheckUser(th.user, username)
	if !errHttp.IsEmpty() {
		log.Info(errHttp)
		log.Info(errHttp.Reason)
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение")
	tenders, err := th.tender.GetListForUser(username, limit, offset)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	tendersDto := []tenderDto{}
	for _, tender := range tenders {
		tendersDto = append(tendersDto, convertToDto(tender))
	}

	c.JSON(http.StatusOK, tendersDto)
}

func (th TenderHandler) getStatusTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckUser(th.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckTender(th.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}
	/*
		log.Info("Авторизация")
		err = auth.CheckUserViewTender(username, tenderId)
		if err == sql.ErrNoRows {
			error.GetUserNotViewTenderError(c)
			return
		} else if err != nil {
			error.GetInternalServerError(c, err)
			return
		}*/

	log.Info("Чтение данных")
	var status string
	err := th.tender.GetStatus(&status, tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}
	c.JSON(http.StatusOK, status)
}

func (th TenderHandler) createTender(c *gin.Context) {
	log.Info("Чтение параметров")
	someTender := db.TenderDefault()
	err := c.BindJSON(&someTender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInvalidRequestFormatOrParametersError(err).SeparateCode())
		return
	}
	log.Info("Валидация")

	errHttp := validator.CheckUser(th.user, someTender.CreatorUsername)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckOrganization(th.organization, someTender.OrganizationId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	/*	log.Info("Авторизация")
		err = auth.CheckUserCanManageTender(someTender.CreatorUsername, someTender.OrganizationId())
		if err == sql.ErrNoRows {
			error.GetUserNotResponsibleOrganizationError(c)
			return
		} else if err != nil {
			error.GetInternalServerError(c, err)
			return
		}
	*/

	log.Info("Создание")
	err = th.tender.Create(&someTender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(someTender))
}

func (th TenderHandler) changeStatusTender(c *gin.Context) {
	log.Info("Чтение параметров")
	status := c.Query("status")
	username := c.Query("username")
	tenderId := c.Param("tenderId")

	log.Info("Валидация")
	if status == "" {
		c.AbortWithStatusJSON(errors.GetNewStatusNotPassedError().SeparateCode())
		return
	}
	if !slices.Contains(StatusConst, status) {
		c.AbortWithStatusJSON(errors.GetInvalidStatusError().SeparateCode())
		return
	}

	errHttp := validator.CheckTender(th.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckUser(th.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных")
	tender := db.TenderDefault()
	err := th.tender.Get(&tender, tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	/*log.Info("Авторизация")
	err = auth.CheckUserCanManageTender(username, tender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}*/

	log.Info("Изменение")
	err = th.tender.ChangeStatus(&status, tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение данных")
	err = th.tender.Get(&tender, tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(tender))
}

func (th TenderHandler) editTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckTender(th.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckUser(th.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных")
	tender := db.TenderDefault()
	err := th.tender.Get(&tender, tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	err = c.BindJSON(&tender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInvalidRequestFormatOrParametersError(err).SeparateCode())
		return
	}
	if !slices.Contains(ServiceTypesConst, tender.ServiceType) {
		c.AbortWithStatusJSON(errors.GetInvalidServiceTypeError().SeparateCode())
		return
	}

	/*log.Info("Авторизация")
	err = auth.CheckUserCanManageTender(username, tender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}*/

	log.Info("Изменение")

	err = th.tender.Edit(&tender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение данных")
	err = th.tender.Get(&tender, tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(tender))
}
func (th TenderHandler) rollbackVersionTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckTender(th.tender, tenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInvalidRequestFormatOrParametersError(err).SeparateCode())
		return
	}

	errHttp = validator.CheckUser(th.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных")
	tender := db.TenderDefault()
	err = th.tender.Get(&tender, tenderId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	/*log.Info("Авторизация")
	err = auth.CheckUserCanManageTender(username, tender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}*/

	if version >= tender.Version {
		c.AbortWithStatusJSON(errors.GetInvalidVersionError().SeparateCode())
		return
	}

	log.Info("Чтение данных")
	var params string
	err = th.tender.GetParamsByVersion(&params, tenderId, version)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	json.Unmarshal([]byte(params), &tender)

	log.Info("Изменение")
	err = th.tender.Edit(&tender)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение данных")
	err = th.tender.Get(&tender, tender.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, convertToDto(tender))
}
