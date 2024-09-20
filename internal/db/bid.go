package db

import (
	_ "embed"

	"github.com/jmoiron/sqlx"
)

type BidModel struct {
	db *sqlx.DB
}

//go:embed queries/checkBidExists.sql
var checkBidExistsQuery string

//go:embed queries/checkUserCanManageBid.sql
var checkUserCanManageBidQuery string

//go:embed queries/checkUserViewBid.sql
var checkUserViewBidQuery string

//go:embed queries/checkUserCanApproveBid.sql
var checkUserCanApproveBidQuery string

func (m BidModel) CheckBidExists(bidId string) error {
	var bidExists bool
	return m.db.Get(&bidExists, checkBidExistsQuery, bidId)
}

func (m BidModel) CheckUserCanManageBid(username, autorType, authorId string) error {
	var canManage bool
	return m.db.Get(&canManage, checkUserCanManageBidQuery, username, authorId, autorType)
}
func (m BidModel) CheckUserViewBid(username, bidId string) error {
	var canView bool
	return m.db.Get(&canView, checkUserViewBidQuery, bidId, username)
}
func (m BidModel) CheckUserCanApproveBid(username, tenderId string) error {
	var canManage bool
	return m.db.Get(&canManage, checkUserCanApproveBidQuery, tenderId, username)
}
