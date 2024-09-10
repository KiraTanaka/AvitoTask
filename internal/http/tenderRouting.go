package http

import (
	"database/sql"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

type tender struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ServiceType     string `json:"serviceType"`
	Status          string `json:"status"`
	Version         int    `json:"version"`
	OrganizationId  string `json:"organizationId"`
	CreatedAt       string `json:"createdAt"`
	CreatorUsername string `json:"creatorUsername"`
}

type InternalErrorBody struct {
	Reason string `json:"reason"`
}

func InitTenderRoutes(routes *gin.RouterGroup, conn *sql.DB) {
	db = conn
	tenderRoutes := routes.Group("/tenders")
	tenderRoutes.GET("/", getTenders)
	tenderRoutes.GET("/my", getUserTender)
	tenderRoutes.GET("/:tender_id/status", getStatusTender)
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
	SELECT uuid,
       name,
	   COALESCE(description,''),	   
       status,
       service_type,
       version,
       created_at
	FROM   tender
	WHERE  service_type = ANY ( $1 )
			OR $2 = 0	
	ORDER BY name
	LIMIT $3 OFFSET $4`

	rows, err := db.Query(query, pq.Array(serviceTypes), len(serviceTypes), limit, offset)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}
	defer rows.Close()

	var tenders []tender
	for rows.Next() {
		var tender tender
		err = rows.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt)
		if err != nil {
			log.Error(err)
			c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
			return
		}
		tenders = append(tenders, tender)
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
	SELECT uuid,
		name,
		COALESCE(description,''),
		status,
		service_type,
		version,
		created_at
	FROM tender
	WHERE creator_user_name = $1
	ORDER BY name
	LIMIT $2 OFFSET $3`

	rows, err := db.Query(query, c.Query("username"), limit, offset)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}
	defer rows.Close()

	var tenders []tender
	for rows.Next() {
		var tender tender
		err = rows.Scan(&tender.Id, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.CreatedAt)
		if err != nil {
			log.Error(err)
			c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
			return
		}
		tenders = append(tenders, tender)
	}
	c.IndentedJSON(http.StatusOK, tenders)
}

func getStatusTender(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	tenderId := c.Param("tender_id")

	var status string
	row := db.QueryRow("Select status from tender where uuid = $1", tenderId)
	err := row.Scan(&status)
	if err != nil {
		log.Error(err)
		c.IndentedJSON(http.StatusInternalServerError, InternalErrorBody{Reason: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, status)
}
