package db

import (
	_ "embed"
	"time"

	"github.com/jmoiron/sqlx"
)

type BidModel struct {
	db *sqlx.DB
}

type Bid struct {
	Id              string  `json:"id" db:"id" binding:"max=100"`
	Name            string  `json:"name" db:"name" binding:"required,max=100"`
	Description     string  `json:"description" db:"description" binding:"required,max=500"`
	Status          string  `json:"status" db:"status" binding:"required,oneof=Created Published Canceled"`
	TenderId        string  `json:"tenderId" db:"tender_id" binding:"required,max=100"`
	AuthorType      string  `json:"authorType" db:"author_type" binding:"required,max=100,oneof=Organization User"`
	AuthorId        string  `json:"authorId" db:"author_id" binding:"required,max=100"`
	Version         int     `json:"version" db:"version" binding:"required,min=1"`
	CreatedAt       string  `json:"createdAt" db:"created_at" binding:"required"`
	Decision        *string `json:"decision" db:"decision"`
	CreatorUsername string  `json:"creatorUsername"`
}

//go:embed queries/bid/checkBidExists.sql
var checkBidExistsQuery string

//go:embed queries/bid/checkUserCanManageBid.sql
var checkUserCanManageBidQuery string

//go:embed queries/bid/checkUserViewBid.sql
var checkUserViewBidQuery string

//go:embed queries/bid/checkUserCanApproveBid.sql
var checkUserCanApproveBidQuery string

//go:embed queries/bid/changeBidDecision.sql
var changeBidDecisionQuery string

//go:embed queries/bid/changeBidStatus.sql
var changeBidStatusQuery string

//go:embed queries/bid/createBid.sql
var createBidQuery string

//go:embed queries/bid/createBidDecision.sql
var createBidDecisionQuery string

//go:embed queries/bid/editBid.sql
var editBidQuery string

//go:embed queries/bid/getBid.sql
var getBidQuery string

//go:embed queries/bid/getBidApprovedCount.sql
var getBidApprovedCountQuery string

//go:embed queries/bid/getBidDecisionCountByUser.sql
var getBidDecisionCountByUserQuery string

//go:embed queries/bid/getBidsByTender.sql
var getBidsByTenderQuery string

//go:embed queries/bid/getBidStatus.sql
var getBidStatusQuery string

//go:embed queries/bid/getParamsBidByVersion.sql
var getParamsBidByVersionQuery string

//go:embed queries/bid/getUserBids.sql
var getUserBidsQuery string

func BidDefault() Bid {
	return Bid{Version: 1,
		CreatedAt: time.Now().Format(time.RFC3339),
		Status:    "Created"}
}

func (m BidModel) BeginTransaction() (*sqlx.Tx, error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (m BidModel) CheckExists(bidId string) error {
	var bidExists bool
	return m.db.Get(&bidExists, checkBidExistsQuery, bidId)
}

func (m BidModel) CheckUserCanManageBid(username, autorType, authorId string) error {
	var canManage bool
	return m.db.Get(&canManage, checkUserCanManageBidQuery, username, authorId, autorType)
}
func (m BidModel) CheckUserView(username, bidId string) error {
	var canView bool
	return m.db.Get(&canView, checkUserViewBidQuery, bidId, username)
}
func (m BidModel) CheckUserCanApproveBid(username, tenderId string) error {
	var canManage bool
	return m.db.Get(&canManage, checkUserCanApproveBidQuery, tenderId, username)
}

func (m BidModel) ChangeDecision(tx *sqlx.Tx, decision, bidId string) error {
	_, err := tx.Exec(changeBidDecisionQuery, decision, bidId)
	if err != nil {
		return err
	}
	return nil
}

func (m BidModel) ChangeStatus(status *string, bidId string) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(changeBidStatusQuery, status, bidId)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (m BidModel) Create(bid *Bid) error {
	var lastInsertId string
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = tx.QueryRow(createBidQuery, bid.Name, bid.Description, bid.Status,
		bid.TenderId, bid.AuthorType, bid.AuthorId,
		bid.Version, bid.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		return err
	}
	tx.Commit()
	bid.Id = lastInsertId
	return nil
}

func (m BidModel) CreateDecision(tx *sqlx.Tx, bidId, username, decision string) error {
	_, err := tx.Exec(createBidDecisionQuery, bidId, username, decision)
	if err != nil {
		return err
	}
	return nil
}

func (m BidModel) Edit(bid *Bid) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExec(editBidQuery, bid)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (m BidModel) Get(bid *Bid, bidId string) error {
	err := m.db.Get(bid, getBidQuery, bidId)
	return err
}

func (m BidModel) GetApprovedCount(tx *sqlx.Tx, decisionCnt *int, bidId string) error {
	err := tx.Get(decisionCnt, getBidApprovedCountQuery, bidId)
	return err
}

func (m BidModel) GetDecisionCountByUser(decisionCnt *int, bidId, username string) error {
	err := m.db.Get(decisionCnt, getBidDecisionCountByUserQuery, bidId, username)
	return err
}

func (m BidModel) GetBidsByTender(tenderId, limit, offset string) ([]Bid, error) {
	bids := []Bid{}
	err := m.db.Select(&bids, getBidsByTenderQuery, tenderId, limit, offset)
	return bids, err
}

func (m BidModel) GetStatus(status *string, bidId string) error {
	err := m.db.Get(status, getBidStatusQuery, bidId)
	return err
}

func (m BidModel) GetParamsByVersion(params *string, bidId string, version int) error {
	err := m.db.Get(params, getParamsBidByVersionQuery, bidId, version)
	return err
}

func (m BidModel) GetListForUser(username, limit, offset string) ([]Bid, error) {
	bids := []Bid{}
	err := m.db.Select(&bids, getUserBidsQuery, username, limit, offset)
	return bids, err
}
