package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var db *sqlx.DB

type tender struct {
	Id              string `json:"id" db:"id" binding:"max=100"`
	Name            string `json:"name" db:"name" binding:"required,max=100"`
	Description     string `json:"description" db:"description" binding:"required,max=500"`
	ServiceType     string `json:"serviceType" db:"service_type" binding:"required,oneof=Construction Delivery Manufacture"`
	Status          string `json:"status" db:"status" binding:"required,oneof=Created Published Closed"`
	Version         int    `json:"version" db:"version" binding:"required,min=1"`
	OrganizationId  string `json:"organizationId" db:"organization_id" binding:"required,max=100"`
	CreatedAt       string `json:"createdAt" db:"created_at" binding:"required"`
	CreatorUsername string `json:"creatorUsername" db:"creator_username"`
}

type InternalErrorBody struct {
	Reason string `json:"reason"`
}

func InitTenderRoutes(routes *gin.RouterGroup, conn *sqlx.DB) {
	db = conn
	tenderRoutes := routes.Group("/tenders")
	//GET
	tenderRoutes.GET("/", getTenders)
	tenderRoutes.GET("/my", getUserTender)
	tenderRoutes.GET("/:tender_id/status", getStatusTender)
	//POST
	tenderRoutes.POST("/new", createTender)
	//PUT
	tenderRoutes.PUT("/:tender_id/status", changeStatusTender)
	tenderRoutes.PUT("/:tender_id/rollback/:version", rollbackVersionTEnder)
	//PATCH
	tenderRoutes.PATCH("/:tender_id/edit", editTender)

}

func getTenders(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	limit := c.Query("limit")
	if limit == "" {
		limit = "5"
	}
	offset := c.Query("offset")
	if offset == "" {
		offset = "0"
	}
	serviceTypes := c.QueryArray("service_type")

	query := `
		SELECT id,
	       name,
		   COALESCE(description,'') as description,
	       status,
	       service_type,
	       version,
	       created_at
		FROM   tender
		WHERE  service_type = ANY ( $1 )
				OR $2 = 0
		ORDER BY name
		LIMIT $3 OFFSET $4`

	var tenders []tender
	err := db.Select(&tenders, query, pq.Array(serviceTypes), len(serviceTypes), limit, offset)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, tenders)
}

func getUserTender(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	limit := c.Query("limit")
	if limit == "" {
		limit = "5"
	}
	offset := c.Query("offset")
	if offset == "" {
		offset = "0"
	}

	query := `
		SELECT id,
			name,
			COALESCE(description,'') as description,
			status,
			service_type,
			version,
			created_at
		FROM tender
		WHERE creator_username = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`
	var tenders []tender
	err := db.Select(&tenders, query, c.Query("username"), limit, offset)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, tenders)
}

func getStatusTender(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	tenderId := c.Param("tender_id")

	var status string
	err := db.Get(&status, "SELECT status FROM tender WHERE id = $1", tenderId)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, status)
}

func createTender(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var someTender tender = tender{Version: 1, CreatedAt: time.RFC3339}
	err := c.BindJSON(&someTender)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, InternalErrorBody{Reason: err.Error()})
		return
	}

	query := `INSERT INTO tender
							(name,
							description,
							service_type,
							status,
							organization_id,
							creator_username)
				VALUES     (:name,
							:description,
							:service_type,
							:status,
							:organization_id,
							:creator_username)`
	tx := db.MustBegin()
	resultDbQuery, err := tx.NamedExec(query, &someTender)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}

	tx.Commit()
	log.Info("RowsAffected and Error after comand:")
	log.Info(resultDbQuery.RowsAffected())
	log.Info(resultDbQuery.LastInsertId())

	c.IndentedJSON(http.StatusOK, someTender)
}

func changeStatusTender(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: fmt.Sprintf("%v", r)})
			return
		}
	}()
	c.Header("Content-Type", "application/json")

	status := c.Query("status")
	username := c.Query("username")
	tenderId := c.Param("tender_id")

	if status == "" || username == "" || tenderId == "" {
		messageError := "Переданы не все обязательные параметры(status, username, tender_id)"
		log.Error(messageError)
		c.IndentedJSON(http.StatusBadRequest, InternalErrorBody{Reason: messageError})
		return
	}

	tx := db.MustBegin()
	resultDbQuery := tx.MustExec("UPDATE tender SET status = $1 WHERE id = $2", c.Query("status"), c.Param("tender_id"))
	log.Info("RowsAffected and Error after comand:")
	log.Info(resultDbQuery.RowsAffected())
	tx.Commit()

	c.IndentedJSON(http.StatusOK, status)
}

func editTender(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	tenderId := c.Param("tender_id")

	var tender tender
	err := db.Get(&tender, "SELECT * FROM tender WHERE id = $1", tenderId)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusBadRequest, InternalErrorBody{Reason: err.Error()})
		return
	}

	err = c.BindJSON(&tender)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
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
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}
	tx.Commit()

	c.IndentedJSON(http.StatusOK, tender)
}
func rollbackVersionTEnder(c *gin.Context) {
	nameFunc := "rollbackVersionTEnder"
	c.Header("Content-Type", "application/json")
	tender := tender{Id: c.Param("tender_id")}
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		log.Error(nameFunc+"(converted to type int): ", err)
		c.IndentedJSON(http.StatusBadRequest, InternalErrorBody{Reason: err.Error()})
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
		log.Error(nameFunc+"(select tender): ", err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}
	if version >= tender.Version {
		messageError := "Указанная версия больше или равна текущей версии тендера"
		log.Error(nameFunc+": ", messageError)
		c.IndentedJSON(http.StatusBadRequest, InternalErrorBody{messageError})
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
		log.Error(nameFunc+"(select params of tender version): ", err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}
	json.Unmarshal([]byte(params), &tender)
	_, err = tx.NamedExec(query, &tender)
	if err != nil {
		log.Error(nameFunc, "(update tender): ", err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}

	tx.Commit()

	c.IndentedJSON(http.StatusOK, tender)
}
