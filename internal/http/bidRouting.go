package http

import (
	"net/http"

	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	db "avitoTask/internal/db"
	"avitoTask/internal/errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type BidHandler struct {
	bid    db.BidModel
	tender db.TenderModel
	user   db.UserModel
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

var BidStatusConst []string = []string{"Created", "Published", "Canceled"}
var BidAuthorType []string = []string{"Organization", "User"}
var BidDecisionType []string = []string{"Approved", "Rejected"}

const Quorum int = 3

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
	bids, err := h.bid.GetUserBids(username, limit, offset)
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

/*
func (h BidHandler) createBid(c *gin.Context) {
	log.Info("Чтение параметров")
	someBid := bid{Version: 1, CreatedAt: time.Now().Format(time.RFC3339), Status: "Created"}
	err := c.BindJSON(&someBid)
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}

	log.Info("Валидация")
	if err := uuid.Validate(someBid.TenderId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	if err := uuid.Validate(someBid.AuthorId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err = validator.CheckUserExists(someBid.CreatorUsername)
	if err == sql.ErrNoRows {
		error.GetUserNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	err = validator.CheckTenderExists(someBid.TenderId)
	if err == sql.ErrNoRows {
		error.GetTenderNotFoundError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanManageBid(someBid.CreatorUsername, someBid.AuthorType, someBid.AuthorId)
	if err == sql.ErrNoRows {
		error.GetUserNotAuthorOrResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Создание")
	var lastInsertId string
	tx, err := db.Beginx()
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	defer tx.Rollback()
	query := `INSERT INTO bid
							(name,
							description,
							status,
							tender_id,
							author_type,
							author_id,
							version,
							created_at)
				VALUES     ($1,
							$2,
							$3,
							$4,
							$5,
							$6,
							$7,
							$8)
						RETURNING id`
	err = tx.QueryRow(query, someBid.Name, someBid.Description, someBid.Status,
		someBid.TenderId, someBid.AuthorType, someBid.AuthorId,
		someBid.Version, someBid.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()
	someBid.Id = lastInsertId

	c.JSON(http.StatusOK, someBid.convertToDto())
}

func (h BidHandler) changeStatusBid(c *gin.Context) {
	log.Info("Чтение параметров")

	status := c.Query("status")
	username := c.Query("username")
	bidId := c.Param("id")

	log.Info("Валидация")
	if status == "" {
		error.GetNewStatusNotPassedError(c)
		return
	}
	if !slices.Contains(BidStatusConst, status) {
		error.GetInvalidStatusError(c)
		return
	}

	if bidId == "" {
		error.GetBidIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(bidId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err := validator.CheckBidExists(bidId)
	if err == sql.ErrNoRows {
		error.GetBidNotFoundError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	if username == "" {
		error.GetUserNotPassedError(c)
		return
	}
	err = validator.CheckUserExists(username)
	if err == sql.ErrNoRows {
		error.GetUserNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Чтение данных")
	bid := bid{}
	err = db.Get(&bid, `SELECT id,
								name,
								status,
								tender_id,
								author_type,
								author_id,
								version,
								created_at
							FROM bid WHERE id = $1`, bidId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanManageBid(username, bid.AuthorType, bid.AuthorId)
	if err == sql.ErrNoRows {
		error.GetUserNotAuthorOrResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Изменение")
	tx, err := db.Beginx()
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE bid SET status = $1 WHERE id = $2", status, bid.Id)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	log.Info("Чтение данных")
	err = db.Get(&bid, `SELECT id,
								name,
								status,
								tender_id,
								author_type,
								author_id,
								version,
								created_at
							FROM bid WHERE id = $1`, bid.Id)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, bid.convertToDto())
}

func (h BidHandler) editBid(c *gin.Context) {
	log.Info("Чтение параметров")
	bidId := c.Param("id")
	username := c.Query("username")

	log.Info("Валидация")
	if bidId == "" {
		error.GetTenderIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(bidId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err := validator.CheckBidExists(bidId)
	if err == sql.ErrNoRows {
		error.GetBidNotFoundError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	if username == "" {
		error.GetUserNotPassedError(c)
		return
	}
	err = validator.CheckUserExists(username)
	if err == sql.ErrNoRows {
		error.GetUserNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Чтение данных")
	bid := bid{}
	err = db.Get(&bid, `SELECT id,
								name,
								status,
								tender_id,
								author_type,
								author_id,
								version,
								created_at
							FROM bid WHERE id = $1`, bidId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	err = c.BindJSON(&bid)
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanManageBid(username, bid.AuthorType, bid.AuthorId)
	if err == sql.ErrNoRows {
		error.GetUserNotAuthorOrResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Изменение")
	query := `UPDATE bid
				SET    name = :name,
						description = :description
				WHERE  id = :id`

	tx, err := db.Beginx()
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	defer tx.Rollback()

	_, err = tx.NamedExec(query, bid)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	log.Info("Чтение данных")
	err = db.Get(&bid, `SELECT id,
								name,
								status,
								tender_id,
								author_type,
								author_id,
								version,
								created_at
							FROM bid WHERE id = $1`, bid.Id)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, bid.convertToDto())
}
func (h BidHandler) rollbackVersionBid(c *gin.Context) {
	log.Info("Чтение параметров")
	bidId := c.Param("id")
	username := c.Query("username")

	log.Info("Валидация")
	if bidId == "" {
		error.GetBidIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(bidId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err := validator.CheckBidExists(bidId)
	if err == sql.ErrNoRows {
		error.GetBidNotFoundError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}

	if username == "" {
		error.GetUserNotPassedError(c)
		return
	}
	err = validator.CheckUserExists(username)
	if err == sql.ErrNoRows {
		error.GetUserNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Чтение данных")
	bid := bid{}
	err = db.Get(&bid, `SELECT id,
								name,
								status,
								tender_id,
								author_type,
								author_id,
								version,
								created_at
							FROM bid WHERE id = $1`, bidId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanManageBid(username, bid.AuthorType, bid.AuthorId)
	if err == sql.ErrNoRows {
		error.GetUserNotAuthorOrResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	if version >= bid.Version {
		error.GetInvalidVersionError(c)
		return
	}

	log.Info("Чтение данных")
	var params string
	err = db.Get(&params, `SELECT params
							FROM bid_version_hist
							WHERE bid_id = $1 AND version = $2`, bid.Id, version)
	if err != nil {
		error.GetVersionNotFoundError(c)
		return
	}
	json.Unmarshal([]byte(params), &bid)

	log.Info("Изменение")
	query := `UPDATE bid
				SET    name = :name,
						description = :description
				WHERE  id = :id`

	tx := db.MustBegin()
	_, err = tx.NamedExec(query, &bid)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	log.Info(bid.Id)

	log.Info("Чтение данных")
	err = db.Get(&bid, `SELECT id,
								name,
								status,
								tender_id,
								author_type,
								author_id,
								version,
								created_at
							FROM bid WHERE id = $1`, bid.Id)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, bid.convertToDto())
}

// Расширенный процесс согласования
func (h BidHandler) SubmitDecisionBid(c *gin.Context) {
	log.Info("Чтение параметров")
	bidId := c.Param("id")
	username := c.Query("username")
	decision := c.Query("decision")

	log.Info("Валидация")
	if decision == "" {
		error.GetDecisionNotPassedError(c)
		return
	}
	if !slices.Contains(BidDecisionType, decision) {
		error.GetInvalidDecisionError(c)
		return
	}

	if bidId == "" {
		error.GetBidIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(bidId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err := validator.CheckBidExists(bidId)
	if err == sql.ErrNoRows {
		error.GetBidNotFoundError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	if username == "" {
		error.GetUserNotPassedError(c)
		return
	}
	err = validator.CheckUserExists(username)
	if err == sql.ErrNoRows {
		error.GetUserNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Чтение данных")
	bid := bid{}
	err = db.Get(&bid, `SELECT id,
								name,
								status,
								tender_id,
								author_type,
								author_id,
								version,
								created_at,
								decision
							FROM bid WHERE id = $1`, bidId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	if bid.Decision != nil {
		error.GetBidAlreadyHasDecisionError(c)
		return
	}

	var decisionCnt int
	err = db.Get(&decisionCnt, `SELECT COUNT(*)
							FROM bid_decision
							WHERE bid_id = $1 AND username=$2`,
		bid.Id, username)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	if decisionCnt >= 1 {
		error.GetUserHasDecisionForBidError(c)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanApproveBid(username, bid.TenderId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Изменение")
	var lastInsertId string
	tx, err := db.Beginx()
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	defer tx.Rollback()
	err = tx.QueryRow(`INSERT INTO bid_decision
									(bid_id,
									username,
									decision)
						VALUES     ($1,
									$2,
									$3)
						RETURNING id`, bid.Id,
		username, decision).Scan(&lastInsertId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	if decision == "Rejected" {

		_, err = tx.Exec("UPDATE bid SET decision = $1 WHERE id = $2", decision, bid.Id)
		if err != nil {
			error.GetInternalServerError(c, err)
			return
		}
	} else {
		err = tx.Get(&decisionCnt, `SELECT COUNT(*)
							FROM bid_decision
							WHERE bid_id = $1 AND decision = 'Approved'`,
			bid.Id)
		if err != nil {
			error.GetInternalServerError(c, err)
			return
		}
		log.Info(decisionCnt)
		if decisionCnt >= Quorum {
			_, err = tx.Exec("UPDATE bid SET decision = $1 WHERE id = $2", decision, bid.Id)
			if err != nil {
				error.GetInternalServerError(c, err)
				return
			}
			_, err = tx.Exec("UPDATE tender SET status = $1 WHERE id = $2", "Closed", bid.TenderId)
			if err != nil {
				error.GetInternalServerError(c, err)
				return
			}
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, bid.convertToDto())
}
*/
