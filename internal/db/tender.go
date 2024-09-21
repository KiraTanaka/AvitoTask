package db

import (
	_ "embed"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Tender struct {
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

type TenderModel struct {
	db *sqlx.DB
}

//go:embed queries/tender/checkTenderExists.sql
var checkTenderExistsQuery string

//go:embed queries/tender/checkUserCanManageTender.sql
var checkUserCanManageTenderQuery string

//go:embed queries/tender/checkUserViewTender.sql
var checkUserViewTenderQuery string

//go:embed queries/tender/getTendersByTypeOfService.sql
var getTendersByTypeOfServiceQuery string

//go:embed queries/tender/getUserTenders.sql
var getUserTendersQuery string

//go:embed queries/tender/getStatusTender.sql
var getStatusTenderQuery string

//go:embed queries/tender/createTender.sql
var createTenderQuery string

//go:embed queries/tender/getTender.sql
var getTenderQuery string

//go:embed queries/tender/changeStatusTender.sql
var changeStatusTenderQuery string

//go:embed queries/tender/editTender.sql
var editTenderQuery string

//go:embed queries/tender/getParamsTenderByVersion.sql
var getParamsTenderByVersionQuery string

func TenderDefault() Tender {
	return Tender{Version: 1,
		CreatedAt: time.Now().Format(time.RFC3339),
		Status:    "Created"}
}

func (m TenderModel) CheckExists(tenderId string) error {
	var tenderExists bool
	return m.db.Get(&tenderExists, checkTenderExistsQuery, tenderId)
}

func (m TenderModel) CheckUserCanManage(username, organizationId string) error {
	var isResponsibleOrganization bool
	return m.db.Get(&isResponsibleOrganization, checkUserCanManageTenderQuery, organizationId, username)
}
func (m TenderModel) CheckUserView(username, tenderId string) error {
	var canView bool
	return m.db.Get(&canView, checkUserViewTenderQuery, tenderId, username)
}
func (m TenderModel) GetListByTypeOfService(serviceTypes []string, limit string, offset string) ([]Tender, error) {
	tenders := []Tender{}
	err := m.db.Select(&tenders, getTendersByTypeOfServiceQuery, pq.Array(serviceTypes), len(serviceTypes), limit, offset)
	return tenders, err
}

func (m TenderModel) GetListForUser(username, limit, offset string) ([]Tender, error) {
	tenders := []Tender{}
	err := m.db.Select(&tenders, getUserTendersQuery, username, limit, offset)
	return tenders, err
}

func (m TenderModel) GetStatus(status *string, tenderId string) error {
	err := m.db.Get(status, getStatusTenderQuery, tenderId)
	return err
}

func (m TenderModel) Get(tender *Tender, tenderId string) error {
	err := m.db.Get(tender, getTenderQuery, tenderId)
	return err
}

func (m TenderModel) Create(tender *Tender) error {
	var lastInsertId string
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = tx.QueryRow(createTenderQuery, tender.Name, tender.Description, tender.ServiceType,
		tender.Status, tender.OrganizationId,
		tender.Version, tender.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		return err
	}
	tx.Commit()
	tender.Id = lastInsertId
	return nil
}

func (m TenderModel) Edit(tender *Tender) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExec(editTenderQuery, tender)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (m TenderModel) ChangeStatus(status *string, tenderId string) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(changeStatusTenderQuery, status, tenderId)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (m TenderModel) GetParamsByVersion(params *string, tenderId string, version int) error {
	err := m.db.Get(params, getParamsTenderByVersionQuery, tenderId, version)
	return err
}
