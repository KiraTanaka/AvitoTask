package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"time"

	validator "avitoTask/internal"
	"avitoTask/internal/auth"
	"avitoTask/internal/error"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

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
}
type tenderDto struct {
	Id          string `json:"id" db:"id" binding:"max=100"`
	Name        string `json:"name" db:"name" binding:"required,max=100"`
	Description string `json:"description" db:"description" binding:"required,max=500"`
	ServiceType string `json:"serviceType" db:"service_type" binding:"required,oneof=Construction Delivery Manufacture"`
	Status      string `json:"status" db:"status" binding:"required,oneof=Created Published Closed"`
	Version     int    `json:"version" db:"version" binding:"required,min=1"`
	CreatedAt   string `json:"createdAt" db:"created_at" binding:"required"`
}

var StatusConst []string = []string{"Created", "Published", "Closed"}
var ServiceTypesConst []string = []string{"Construction", "Delivery", "Manufacture"}

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

func (t *tender) convertToDto() *tenderDto {
	var tenderDto tenderDto
	tenderDto.Id = t.Id
	tenderDto.Name = t.Name
	tenderDto.Description = t.Description
	tenderDto.ServiceType = t.ServiceType
	tenderDto.Status = t.Status
	tenderDto.Version = t.Version
	tenderDto.CreatedAt = t.CreatedAt
	return &tenderDto
}

func getTenders(c *gin.Context) {
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
			error.GetInvalidServiceTypeError(c)
			return
		}

	}

	log.Info("Чтение")
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

	var tenders []tenderDto
	err := db.Select(&tenders, query, pq.Array(serviceTypes), len(serviceTypes), limit, offset)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func getUserTender(c *gin.Context) {
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

	log.Info("Чтение")
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
	var tenders []tenderDto
	err = db.Select(&tenders, query, username, limit, offset)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func getStatusTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	if tenderId == "" {
		error.GetTenderIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(tenderId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
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

	log.Info("Авторизация")
	err = auth.CheckUserViewTender(username, tenderId)
	if err == sql.ErrNoRows {
		error.GetUserNotViewTenderError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Чтение данных")
	var status string
	err = db.Get(&status, "SELECT status FROM tender WHERE id = $1", tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}

func createTender(c *gin.Context) {
	log.Info("Чтение параметров")
	someTender := tender{Version: 1, CreatedAt: time.Now().Format(time.RFC3339), Status: "Created"}
	err := c.BindJSON(&someTender)
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	log.Info("Валидация")
	if err := uuid.Validate(someTender.OrganizationId); err != nil {
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

	log.Info("Авторизация")
	err = auth.CheckUserCanManageTender(someTender.CreatorUsername, someTender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
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
	err = tx.QueryRow(`INSERT INTO tender
									(name,
									description,
									service_type,
									status,
									organization_id,
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
						RETURNING id`, someTender.Name, someTender.Description, someTender.ServiceType, someTender.Status, someTender.OrganizationId,
		someTender.Version, someTender.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()
	someTender.Id = lastInsertId

	c.JSON(http.StatusOK, someTender.convertToDto())
}

func changeStatusTender(c *gin.Context) {
	log.Info("Чтение параметров")
	status := c.Query("status")
	username := c.Query("username")
	tenderId := c.Param("tenderId")

	log.Info("Валидация")
	if status == "" {
		error.GetNewStatusNotPassedError(c)
		return
	}
	if !slices.Contains(StatusConst, status) {
		error.GetInvalidStatusError(c)
		return
	}

	if tenderId == "" {
		error.GetTenderIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(tenderId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err := validator.CheckTenderExists(tenderId)
	if err == sql.ErrNoRows {
		error.GetTenderNotFoundError(c)
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
	var tender tender
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								organization_id,
								version,
								created_at
							FROM tender WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanManageTender(username, tender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
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
	_, err = tx.Exec("UPDATE tender SET status = $1 WHERE id = $2", status, tender.Id)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	log.Info("Чтение данных")
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								organization_id,
								version,
								created_at
							FROM tender
							WHERE id = $1`, tender.Id)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tender.convertToDto())
}

func editTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	if tenderId == "" {
		error.GetTenderIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(tenderId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err := validator.CheckTenderExists(tenderId)
	if err == sql.ErrNoRows {
		error.GetTenderNotFoundError(c)
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
	var tender tender
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								organization_id,
								version,
								created_at
							FROM tender WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	err = c.BindJSON(&tender)
	if err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	if !slices.Contains(ServiceTypesConst, tender.ServiceType) {
		error.GetInvalidServiceTypeError(c)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanManageTender(username, tender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Изменение")
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

	log.Info("Чтение данных")
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								organization_id,
								version,
								created_at
							FROM tender
							WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tender.convertToDto())
}
func rollbackVersionTender(c *gin.Context) {
	log.Info("Чтение параметров")
	tenderId := c.Param("tenderId")
	username := c.Query("username")

	log.Info("Валидация")
	if tenderId == "" {
		error.GetTenderIdNotPassedError(c)
		return
	}
	if err := uuid.Validate(tenderId); err != nil {
		error.GetInvalidRequestFormatOrParametersError(c, err)
		return
	}
	err := validator.CheckTenderExists(tenderId)
	if err == sql.ErrNoRows {
		error.GetTenderNotFoundError(c)
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
	var tender tender
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								organization_id,
								version,
								created_at
							FROM tender WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	log.Info("Авторизация")
	err = auth.CheckUserCanManageTender(username, tender.OrganizationId)
	if err == sql.ErrNoRows {
		error.GetUserNotResponsibleOrganizationError(c)
		return
	} else if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	if version >= tender.Version {
		error.GetInvalidVersionError(c)
		return
	}

	log.Info("Чтение данных")
	var params string
	err = db.Get(&params, `SELECT params 
							FROM tender_version_hist 
							WHERE tender_id = $1 AND version = $2`, tender.Id, version)
	if err != nil {
		error.GetVersionNotFoundError(c)
		return
	}
	json.Unmarshal([]byte(params), &tender)

	log.Info("Изменение")
	query := `UPDATE tender
				SET    name = :name,
						description = :description,
						service_type = :service_type
				WHERE  id = :id`

	tx := db.MustBegin()
	_, err = tx.NamedExec(query, &tender)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}
	tx.Commit()

	log.Info("Чтение данных")
	err = db.Get(&tender, `SELECT id,
								name,
								COALESCE(description,'') as description,
								status,
								service_type,
								organization_id,
								version,
								created_at
							FROM tender
							WHERE id = $1`, tenderId)
	if err != nil {
		error.GetInternalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, tender.convertToDto())
}
