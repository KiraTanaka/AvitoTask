package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	"avitoTask/internal/error"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type tender struct {
	Id              string `json:"id" db:"id" binding:"max=100,uuid4"`
	Name            string `json:"name" db:"name" binding:"required,max=100"`
	Description     string `json:"description" db:"description" binding:"required,max=500"`
	ServiceType     string `json:"serviceType" db:"service_type" binding:"required,oneof=Construction Delivery Manufacture"`
	Status          string `json:"status" db:"status" binding:"required,oneof=Created Published Closed"`
	Version         int    `json:"version" db:"version" binding:"required,min=1"`
	OrganizationId  string `json:"organizationId" db:"organization_id" binding:"required,max=100,uuid4"`
	CreatedAt       string `json:"createdAt" db:"created_at" binding:"required"`
	CreatorUsername string `json:"creatorUsername" db:"creator_username"`
}
type tenderDto struct{}

func InitTenderRoutes(routes *gin.RouterGroup) {
	tenderRoutes := routes.Group("/tenders")
	//GET
	tenderRoutes.GET("/", getTenders)
	tenderRoutes.GET("/my", getUserTender)
	tenderRoutes.GET("/:tenderId/status", getStatusTender)
	//POST
	tenderRoutes.POST("/new", createTender)
	//PUT
	tenderRoutes.PUT("/:tenderId/status", changeStatusTender)
	tenderRoutes.PUT("/:tenderId/rollback/:version", rollbackVersionTender)
	//PATCH
	tenderRoutes.PATCH("/:tenderId/edit", editTender)

}

func getTenders(c *gin.Context) {
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
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func getUserTender(c *gin.Context) {
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
			service_type,
			version,
			created_at
		FROM tender
		WHERE creator_username = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`
	var tenders []tender
	err = db.Select(&tenders, query, c.Query("username"), limit, offset)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func getStatusTender(c *gin.Context) {
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	if tenderId == "" {
		error.GetTenderNotFoundError(c)
		return
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

	err = validator.CheckTenderExists(tenderId)
	if err == sql.ErrNoRows {
		error.GetTenderNotFoundError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	var status string
	err = db.Get(&status, "SELECT status FROM tender WHERE id = $1", tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func createTender(c *gin.Context) {
	someTender := tender{Version: 1, CreatedAt: time.RFC3339, Status: "Created"}
	err := c.BindJSON(&someTender)
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}

	err = validator.CheckUserExists(someTender.CreatorUsername)
	if err == sql.ErrNoRows {
		error.GetUserNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	err = validator.CheckOrganizationExists(someTender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetOrganizationNotExistsOrIncorrectError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	//Проверка прав на создание
	isResponsibleOrganization, err := auth.CheckUserIsResponsibleOrganization(someTender.CreatorUsername, someTender.OrganizationId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	} else if !isResponsibleOrganization {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	}

	var lastInsertId string
	tx, err := db.Beginx()
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	defer tx.Rollback()
	err = tx.QueryRow(`INSERT INTO tender
									(name,
									description,
									service_type,
									status,
									organization_id,
									version
									created_at)
						VALUES     ($1,
									$2,
									$3,
									$4,
									$5,
									$6,
									$7)
						RETURNING id`, someTender.Name, someTender.Description, someTender.ServiceType, someTender.Status, someTender.OrganizationId,
		someTender.Version, someTender.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()
	someTender.Id = lastInsertId

	c.JSON(http.StatusOK, someTender)
}

func changeStatusTender(c *gin.Context) {
	status := c.Query("status")
	username := c.Query("username")
	tenderId := c.Param("tenderId")

	if status == "" {
		error.GetNewStatusNotPassedError(c)
		return
	}
	if tenderId == "" {
		error.GetTenderNotFoundError(c)
		return
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

	tx, err := db.Beginx()
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE tender SET status = $1 WHERE id = $2", c.Query("status"), c.Param("tender_id"))
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	var tender tender
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								version,
								created_at
							FROM tender
							WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tender)
}

func editTender(c *gin.Context) {
	tenderId := c.Param("tenderId")
	username := c.Query("username")
	if tenderId == "" {
		error.GetTenderNotFoundError(c)
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

	var tender tender
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								version,
								created_at
							FROM tender WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	//Проверка прав на редактирование
	isResponsibleOrganization, err := auth.CheckUserIsResponsibleOrganization(username, tender.OrganizationId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	} else if !isResponsibleOrganization {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	}

	err = c.BindJSON(&tender)
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}

	query := `UPDATE tender
				SET    name = :name,
						description = :description,
						service_type = :service_type
				WHERE  id = :id`

	tx, err := db.Beginx()
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	defer tx.Rollback()

	_, err = tx.NamedExec(query, tender)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								version,
								created_at
							FROM tender
							WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tender)
}
func rollbackVersionTender(c *gin.Context) {
	tenderId := c.Param("tenderId")
	username := c.Query("username")
	if tenderId == "" {
		error.GetTenderNotFoundError(c)
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

	var tender tender
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								version,
								created_at
							FROM tender WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	} else if tender.Id == "" {
		error.GetTenderNotFoundError(c)
		return
	}

	//Проверка прав на откат
	isResponsibleOrganization, err := auth.CheckUserIsResponsibleOrganization(username, tender.OrganizationId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	} else if !isResponsibleOrganization {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	}

	if version >= tender.Version {
		error.GetInvalidVersionError(c)
		return
	}

	query := `UPDATE tender
				SET    name = :name,
						description = :description,
						service_type = :service_type,
				WHERE  id = :id`

	tx := db.MustBegin()

	var params string
	err = db.Get(&params, `SELECT params 
							FROM tender_version_hist 
							WHERE tender_id = $1 AND version = $2`, tender.Id, version)
	if err != nil {
		error.GetVersionNotFoundError(c)
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
