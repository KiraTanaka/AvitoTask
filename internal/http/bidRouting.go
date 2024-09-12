package http

import (
	"database/sql"
	"net/http"

	validator "avitoTask/internal"
	_ "avitoTask/internal/auth"
	"avitoTask/internal/error"

	"github.com/gin-gonic/gin"
)

type bid struct {
	Id          string `json:"id" db:"id" binding:"max=100,uuid4"`
	Name        string `json:"name" db:"name" binding:"required,max=100"`
	Description string `json:"description" db:"description" binding:"required,max=500"`
	Status      string `json:"status" db:"status" binding:"required,oneof=Created Published Closed"`
	TenderId    string `json:"tenderId" db:"tender_id" binding:"required,max=100,uuid4"`
	AuthorType  string `json:"authorType" db:"author_type" binding:"required,max=100,oneof=Organization User"`
	AuthorId    string `json:"authorId" db:"author_id" binding:"required,max=100,uuid4"`
	Version     int    `json:"version" db:"version" binding:"required,min=1"`
	CreatedAt   string `json:"createdAt" db:"created_at" binding:"required"`
}

func InitBidRoutes(routes *gin.RouterGroup) {
	bidRoutes := routes.Group("/bids")
	//GET
	bidRoutes.GET("/:tenderId/list", getBidsListTender)
	/*bidRoutes.GET("/my", getUserBids)
	bidRoutes.GET("/:bidId/status", getStatusBid)
	//POST
	bidRoutes.POST("/new", createBid)
	//PUT
	bidRoutes.PUT("/:bidId/status", changeStatusBid)
	//bidRoutes.PUT("/:bidId/rollback/:version", rollbackVersionBid)
	//PATCH
	bidRoutes.PATCH("/:bidId/edit", editBid)*/
	/*
		bidRoutes.PUT("/:bidId/submit_decision", SubmitDecisionBid)
		bidRoutes.PUT("/:bidId/feedback", feedbackBid)
		bidRoutes.GET("/:tenderId/reviews", getReviewsOfBid)
	*/

}

func getBidsListTender(c *gin.Context) {
	tenderId := c.Param("tender_id")
	limit := c.Query("limit")
	offset := c.Query("offset")
	username := c.Query("username")

	if limit == "" {
		limit = "5"
	}
	if offset == "" {
		offset = "0"
	}
	if username == "" {
		error.GetUserNotPassedError(c)
		return
	}

	err := validator.CheckUserExists(username)
	if err == sql.ErrNoRows {
		error.GetUserNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	query := `
		SELECT id,
	       name,
		   COALESCE(description,'') as description,
	       status,
		   tender_id,
	       author_type,
		   author_id,
	       version,
	       created_at
		FROM   bid
		WHERE tender_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`

	var bids []bid
	err = db.Select(&bids, query, tenderId, limit, offset)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, bids)
}

/*
func getUserBids(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	limit := c.Query("limit")
	if limit == "" {
		limit = "5"
	}
	offset := c.Query("offset")
	if offset == "" {
		offset = "0"
	}
	username := c.Query("username")
	if username == "" {
		error.GetUserNotPassedError(c)
		return
	}

	query := `
		SELECT id,
	       name,
		   COALESCE(description,'') as description,
	       status,
		   tender_id,
	       author_type,
		   author_id,
	       version,
	       created_at
		FROM bids
		WHERE creator_username = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`
	var bids []bid
	err := db.Select(&bids, query, c.Query("username"), limit, offset)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, bids)
}

func getStatusBid(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	bidId := c.Param("bidId")

	var status string
	err := db.Get(&status, "SELECT status FROM bid WHERE id = $1", bidId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func createBid(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var someBid bid = bid{Version: 1, CreatedAt: time.RFC3339, Status: "Created"}
	err := c.BindJSON(&someBid)
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}

	var objectExist bool
	err = db.Get(&objectExist, "SELECT true FROM tender WHERE id = $1", someBid.TenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	if !objectExist {
		error.GetTenderNotFoundError(c)
		return
	}
	err = db.Get(&objectExist, `SELECT TRUE
								FROM   employee
								WHERE  id = $1
									AND 'User' = $2
								UNION
								SELECT TRUE
								FROM   organization
								WHERE  id = $1
									AND 'Organization' = $2 `, someBid.AuthorId, someBid.AuthorType)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	if !objectExist {
		error.GetAuthorNotFoundError(c)
		return
	}

	query := `INSERT INTO bid
							(name,
							description,
							status,
							tender_id,
							author_type,
							author_id,
							creator_username,
							created_at)
				VALUES     (:name,
							:description,
							:status,
							:tender_id,
							:author_type,
							:author_id,
							:creator_username,
							:created_at)`
	tx := db.MustBegin()
	resultDbQuery, err := tx.NamedExec(query, &someBid)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()
	log.Info(resultDbQuery.LastInsertId())

	c.JSON(http.StatusOK, someBid)
}

func changeStatusBid(c *gin.Context) {
	/*defer func() {
		if r := recover(); r != nil {
			error.GetInternalServerError(c, error.Newfmt.Sprintf("%v", r))
			log.Error(r)
			c.JSON(http.StatusInternalServerError, InternalErrorBody{Reason: })
			return
		}
	}()
	c.Header("Content-Type", "application/json")

	status := c.Query("status")
	username := c.Query("username")
	bidId := c.Param("bidId")

	if status == "" {
		error.GetNewStatusNotPassedError(c)
		return
	}
	if bidId == "" {
		error.GetBidIdNotPassedError(c)
		return
	}
	if username == "" {
		error.GetUserNotPassedError(c)
		return
	}

	tx := db.MustBegin()
	resultDbQuery := tx.MustExec("UPDATE tender SET status = $1 WHERE id = $2", c.Query("status"), c.Param("bidId"))
	log.Info("RowsAffected and Error after comand:")
	log.Info(resultDbQuery.RowsAffected())
	tx.Commit()

	c.JSON(http.StatusOK, status)
}

func editBid(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	tenderId := c.Param("tender_id")

	var tender tender
	err := db.Get(&tender, "SELECT * FROM tender WHERE id = $1", tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	err = c.BindJSON(&tender)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	query := `
	UPDATE tender
	SET    name = :name,
			description = :description,
			service_type = :service_type,
			status = :status
	WHERE  id = :id`

	tx := db.MustBegin()
	resultDbQuery, err := tx.NamedExec(query, tender)
	log.Info("RowsAffected and Error after comand:")
	log.Info(resultDbQuery.RowsAffected())
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, tender)
}*/ /*
func rollbackVersionTEnder(c *gin.Context) {
	nameFunc := "rollbackVersionTEnder"
	c.Header("Content-Type", "application/json")
	tender := tender{Id: c.Param("tender_id")}
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		log.Error(nameFunc+"(converted to type int): ", err)
		error.GetInternalServerError(c, err)
		//c.JSON(http.StatusBadRequest, InternalErrorBody{Reason: err.Error()})
		return
	}

	query := `SELECT  name,
					COALESCE(description,'') as description,
					status,
					service_type,
					version
			FROM tender WHERE id = $1`

	err = db.Get(&tender, query, tender.Id)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	if version >= tender.Version {
		error.GetInternalServerError(c, err)
		return
	}

	query = `
	UPDATE tender
	SET    name = :name,
			description = :description,
			service_type = :service_type,
			status = :status
	WHERE  id = :id`

	tx := db.MustBegin()

	var params string
	err = db.Get(&params, "SELECT params FROM tender_version_hist2 WHERE tender_id = $1 and version = $2", tender.Id, version)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	json.Unmarshal([]byte(params), &tender)
	_, err = tx.NamedExec(query, &tender)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, tender)
}
*/
