package http

import (
	"encoding/json"
	"net/http"

	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	db "avitoTask/internal/db"
	"avitoTask/internal/errors"
	"avitoTask/internal/services"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type BidHandler struct {
	bid          db.BidModel
	tender       db.TenderModel
	user         db.UserModel
	organization db.OrganizationModel
}

type bidDto struct {
	Id         string `json:"id" db:"id" binding:"max=100"`
	Name       string `json:"name" db:"name" binding:"required,max=100"`
	Status     string `json:"status" db:"status" binding:"required,oneof=Created Published Closed"`
	AuthorType string `json:"authorType" db:"author_type" binding:"required,max=100,oneof=Organization User"`
	AuthorId   string `json:"authorId" db:"author_id" binding:"required,max=100"`
	Version    int    `json:"version" db:"version" binding:"required,min=1"`
	CreatedAt  string `json:"createdAt" db:"created_at" binding:"required"`
}

type bidDecision struct {
	Id       string `json:"id" db:"id" binding:"max=100"`
	BidId    string `json:"bidId" db:"bid_id" binding:"max=100"`
	Username string `json:"username" db:"username" binding:"max=50"`
	Decision string `json:"decision" db:"decision" binding:"oneof=Approved Rejected"`
}

var bidStatusesConst []string = []string{"Created", "Published", "Canceled"}
var bidAuthorTypesConst []string = []string{"Organization", "User"}
var bidDecisionTypesConst []string = []string{"Approved", "Rejected"}

func InitBidRoutes(routes *gin.RouterGroup, bidHandler *BidHandler) {
	bidRoutes := routes.Group("/bids")
	//GET
	bidRoutes.GET("/:id/list", bidHandler.getBidsListTender)
	bidRoutes.GET("/my", bidHandler.getUserBids)
	bidRoutes.GET("/:id/status", bidHandler.getStatusBid)
	/*//POST
	bidRoutes.POST("/new", createBid)
	//PUT
	bidRoutes.PUT("/:id/status", changeStatusBid)
	bidRoutes.PUT("/:id/rollback/:version", rollbackVersionBid)
	bidRoutes.PUT("/:id/submit_decision", SubmitDecisionBid)
	//PATCH
	bidRoutes.PATCH("/:id/edit", editBid)*/
	//	bidRoutes.PUT("/:bidId/feedback", feedbackBid)
	//	bidRoutes.GET("/:tenderId/reviews", getReviewsOfBid)

}

func bidConvertToDto(t db.Bid) bidDto {
	var bidDto bidDto
	bidDto.Id = t.Id
	bidDto.Name = t.Name
	bidDto.AuthorType = t.AuthorType
	bidDto.AuthorId = t.AuthorId
	bidDto.Status = t.Status
	bidDto.Version = t.Version
	bidDto.CreatedAt = t.CreatedAt
	return bidDto
}

// По заданию непонятно какие права должны быть
func (h BidHandler) getBidsListTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("id")
	username := c.Query("username")

	limit, offset := SetDefaultPaginationParamIfEmpty(c.Query("limit"), c.Query("offset"))

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

	//По заданию непонятно какие права должны быть

	log.Info("Чтение данных")
	bids, err := h.bid.GetBidsByTender(tenderId, limit, offset)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	bidsDto := []bidDto{}
	for _, bid := range bids {
		bidsDto = append(bidsDto, bidConvertToDto(bid))
	}

	c.JSON(http.StatusOK, bidsDto)
}

func (h BidHandler) getUserBids(c *gin.Context) {
	username := c.Query("username")
	limit, offset := SetDefaultPaginationParamIfEmpty(c.Query("limit"), c.Query("offset"))

	log.Info("Валидация")
	errHttp := validator.CheckUser(h.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение")
	bids, err := h.bid.GetListForUser(username, limit, offset)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	bidsDto := []bidDto{}
	for _, bid := range bids {
		bidsDto = append(bidsDto, bidConvertToDto(bid))
	}

	c.JSON(http.StatusOK, bidsDto)
}

func (h BidHandler) getStatusBid(c *gin.Context) {
	log.Info("Чтение параметров")
	bidId := c.Param("id")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckBid(h.bid, bidId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckUser(h.user, username)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserViewBid(h.bid, username, bidId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных")
	var status string
	err := h.bid.GetStatus(&status, bidId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h BidHandler) createBid(c *gin.Context) {
	log.Info("Чтение параметров")
	someBid := db.BidDefault()
	err := c.BindJSON(&someBid)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInvalidRequestFormatOrParametersError(err).SeparateCode())
		return
	}

	log.Info("Валидация")
	errHttp := validator.CheckTender(h.tender, someBid.TenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.AuthorTypeAcceptable(someBid.AuthorType, bidAuthorTypesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	if someBid.AuthorType == "User" {
		errHttp := validator.CheckUser(h.user, someBid.AuthorId)
		if !errHttp.IsEmpty() {
			c.AbortWithStatusJSON(errHttp.SeparateCode())
			return
		}
	} else if someBid.AuthorType == "Organization" {
		errHttp = validator.CheckOrganization(h.organization, someBid.AuthorId)
		if !errHttp.IsEmpty() {
			c.AbortWithStatusJSON(errHttp.SeparateCode())
			return
		}
	}

	errHttp = validator.CheckUser(h.user, someBid.CreatorUsername)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageBid(h.bid, someBid.CreatorUsername, someBid.AuthorType, someBid.AuthorId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Создание")
	err = h.bid.Create(&someBid)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, bidConvertToDto(someBid))
}

func (h BidHandler) changeStatusBid(c *gin.Context) {
	log.Info("Чтение параметров")

	status := c.Query("status")
	username := c.Query("username")
	bidId := c.Param("id")

	log.Info("Валидация")
	errHttp := validator.CheckStatus(status, bidStatusesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckBid(h.bid, bidId)
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
	bid := db.BidDefault()
	err := h.bid.Get(&bid, bidId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageBid(h.bid, username, bid.AuthorType, bid.AuthorId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Изменение")
	err = h.bid.ChangeStatus(&status, bid.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение измененных данных")
	err = h.bid.Get(&bid, bid.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, bidConvertToDto(bid))
}

func (h BidHandler) editBid(c *gin.Context) {
	log.Info("Чтение параметров")
	bidId := c.Param("id")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckBid(h.bid, bidId)
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
	bid := db.BidDefault()
	err := h.bid.Get(&bid, bidId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение новых значений")
	err = c.BindJSON(&bid)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInvalidRequestFormatOrParametersError(err).SeparateCode())
		return
	}

	errHttp = validator.AuthorTypeAcceptable(bid.AuthorType, bidAuthorTypesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageBid(h.bid, username, bid.AuthorType, bid.AuthorId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Изменение")
	err = h.bid.Edit(&bid)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение измененных данных")
	err = h.bid.Get(&bid, bid.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, bidConvertToDto(bid))
}

func (h BidHandler) rollbackVersionBid(c *gin.Context) {
	log.Info("Чтение параметров")
	bidId := c.Param("id")
	username := c.Query("username")

	log.Info("Валидация")
	errHttp := validator.CheckBid(h.bid, bidId)
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
	bid := db.BidDefault()
	err := h.bid.Get(&bid, bidId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	if version >= bid.Version {
		c.AbortWithStatusJSON(errors.GetVersionIsOutOfBoundsError().SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanManageBid(h.bid, username, bid.AuthorType, bid.AuthorId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Чтение данных версии")
	var params string
	err = h.bid.GetParamsByVersion(&params, bid.Id, version)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	json.Unmarshal([]byte(params), &bid)

	log.Info("Откат до версии")
	err = h.bid.Edit(&bid)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	log.Info("Чтение измененных данных")
	err = h.bid.Get(&bid, bid.Id)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, bidConvertToDto(bid))
}

// Расширенный процесс согласования
func (h BidHandler) SubmitDecisionBid(c *gin.Context) {
	log.Info("Чтение параметров")
	bidId := c.Param("id")
	username := c.Query("username")
	decision := c.Query("decision")

	log.Info("Валидация")
	errHttp := validator.CheckBidDecision(decision, bidDecisionTypesConst)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	errHttp = validator.CheckBid(h.bid, bidId)
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
	bid := db.BidDefault()
	err := h.bid.Get(&bid, bidId)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	if bid.Decision != nil {
		c.AbortWithStatusJSON(errors.GetBidAlreadyHasDecisionError().SeparateCode())
		return
	}

	var decisionCnt int
	err = h.bid.GetDecisionCountByUser(&decisionCnt, bid.Id, username)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	if decisionCnt >= 1 {
		c.AbortWithStatusJSON(errors.GetUserHasDecisionForBidError().SeparateCode())
		return
	}

	log.Info("Авторизация")
	errHttp = auth.CheckUserCanApproveBid(h.bid, username, bid.TenderId)
	if !errHttp.IsEmpty() {
		c.AbortWithStatusJSON(errHttp.SeparateCode())
		return
	}

	log.Info("Запись решения")
	services.MakingDecision(h.bid, h.tender, bid.Id, bid.TenderId, username, decision)
	if err != nil {
		c.AbortWithStatusJSON(errors.GetInternalServerError(err).SeparateCode())
		return
	}

	c.JSON(http.StatusOK, bidConvertToDto(bid))
}
